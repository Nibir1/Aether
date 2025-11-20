// aether/search.go
//
// This file implements Aether.Search — the high-level retrieval pipeline.
// At Stage 9, Search primarily orchestrates URL-based retrieval using
// previously implemented subsystems:
//
//   - SmartQuery (intent and routing)
//   - Fetch (robots.txt-compliant HTTP)
//   - Detect (content type / HTML vs JSON vs RSS, etc.)
//   - ParseHTML (structural HTML parsing)
//   - ExtractArticleFromHTML (Readability-style article extraction)
//   - ParseRSS (RSS/Atom feed parsing)
//
// For non-URL queries, Search currently focuses on classification and
// routing (via SmartQuery) without executing multi-source network calls.
// Later stages will extend Search to consult multiple open APIs and
// legal search sources for free-form queries.

package aether

import (
	"context"
	"strconv"
	"strings"

	idetect "github.com/Nibir1/Aether/internal/detect"
	irss "github.com/Nibir1/Aether/internal/rss"
)

// SearchDocumentKind describes the kind of document Aether returns
// as a primary result for a Search call.
type SearchDocumentKind string

const (
	SearchDocumentKindUnknown SearchDocumentKind = "unknown"
	SearchDocumentKindArticle SearchDocumentKind = "article"
	SearchDocumentKindHTML    SearchDocumentKind = "html_page"
	SearchDocumentKindFeed    SearchDocumentKind = "feed"
	SearchDocumentKindJSON    SearchDocumentKind = "json"
	SearchDocumentKindText    SearchDocumentKind = "text"
	SearchDocumentKindBinary  SearchDocumentKind = "binary"
)

// SearchDocument is a normalized representation of a single primary
// piece of content discovered during Search.
//
// It is intentionally LLM-friendly and captures enough context for
// downstream reasoning and TOON/JSON modeling.
type SearchDocument struct {
	URL      string
	Kind     SearchDocumentKind
	Title    string
	Content  string // main text content
	HTML     string // sanitized HTML fragment where applicable
	Excerpt  string
	Source   string            // e.g. "url:article", "url:rss", "url:html", "url:json"
	Metadata map[string]string // arbitrary key/value metadata, if available
}

// SearchResult is the top-level output of Aether.Search.
//
// At Stage 9, SearchResult focuses on URL-oriented retrieval. For
// non-URL queries, it still returns the SmartQueryPlan and routing
// information but may not yet contain network-derived documents.
type SearchResult struct {
	Query           string
	Plan            SmartQueryPlan
	PrimaryDocument *SearchDocument

	// Optional richer views, when applicable:
	Article *Article // when Kind == article
	Feed    *Feed    // when Kind == feed

	// Notes contains human-readable explanations of decisions made
	// by the Search pipeline. These are useful for debugging and
	// introspection.
	Notes []string
}

// Search analyzes the query using SmartQuery and, when appropriate,
// performs a robots.txt-compliant retrieval of a single URL, passing
// the response through Aether's detection, parsing, extraction and
// normalization layers.
//
// Behavior at Stage 9:
//   - If the query is a single URL, Search will fetch and analyze it.
//   - For non-URL queries, Search returns a routing plan and notes
//     but does not yet perform multi-source network retrieval.
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error) {
	plan := c.SmartQuery(query)

	result := &SearchResult{
		Query:           plan.Query,
		Plan:            *plan,
		PrimaryDocument: nil,
		Article:         nil,
		Feed:            nil,
		Notes:           nil,
	}

	// If the SmartQuery plan indicates the query is a single URL,
	// perform URL-based retrieval.
	if plan.HasURL && len(strings.Fields(plan.Query)) == 1 {
		url := strings.TrimSpace(plan.Query)
		result.Notes = append(result.Notes, "Detected single-URL query; performing URL-based retrieval pipeline.")
		if err := c.searchFromURL(ctx, url, result); err != nil {
			return result, err
		}
		return result, nil
	}

	// No URL detected: at Stage 9, we only return routing info.
	result.Notes = append(result.Notes,
		"No direct URL detected in query; returning SmartQuery routing plan only.",
		"Future stages will execute multi-source retrieval (lookup, RSS, OpenAPIs) for this query.",
	)

	return result, nil
}

// searchFromURL executes the full URL-based pipeline:
//
//	Fetch → Detect → (HTML/Article | RSS | JSON | Text | Binary).
func (c *Client) searchFromURL(ctx context.Context, url string, res *SearchResult) error {
	// Step 1: robots.txt-compliant fetch
	fetchRes, err := c.Fetch(ctx, url)
	if err != nil {
		res.Notes = append(res.Notes, "Fetch failed for URL.")
		return err
	}

	// Step 2: content-type detection (MIME + sniffing).
	d := idetect.Detect(fetchRes.Body, fetchRes.Header)
	res.Notes = append(res.Notes, "Content detection completed for fetched URL.")

	switch d.RawType {
	case idetect.TypeHTML:
		return c.handleHTMLURL(url, fetchRes.Body, d, res)
	case idetect.TypeRSS, idetect.TypeXML:
		// Try to treat as feed first.
		return c.handleFeedURL(url, fetchRes.Body, res)
	case idetect.TypeJSON:
		return c.handleJSONURL(url, fetchRes.Body, res)
	case idetect.TypeText:
		return c.handleTextURL(url, fetchRes.Body, res)
	case idetect.TypeImage, idetect.TypePDF, idetect.TypeBinary:
		return c.handleBinaryURL(url, d.RawType, res)
	default:
		// Unknown type: treat as binary.
		return c.handleBinaryURL(url, idetect.TypeUnknown, res)
	}
}

// handleHTMLURL processes an HTML response: it may run article extraction
// or fall back to structural HTML parsing.
func (c *Client) handleHTMLURL(
	url string,
	body []byte,
	detectRes *idetect.Result,
	res *SearchResult,
) error {
	// Parse HTML to provide metadata as a fallback if article extraction fails.
	parsed, parsedErr := c.ParseHTML(body)

	// Prefer article extraction if the page appears article-like or
	// if subtype is unknown (we can still attempt Readability).
	if detectRes.SubType == idetect.TypeArticle || detectRes.SubType == idetect.TypeUnknown {
		article, err := c.ExtractArticleFromHTML(body, url)
		if err == nil && article != nil && strings.TrimSpace(article.Content) != "" {
			res.Article = article
			doc := &SearchDocument{
				URL:      url,
				Kind:     SearchDocumentKindArticle,
				Title:    article.Title,
				Content:  article.Content,
				HTML:     article.HTML,
				Excerpt:  article.Excerpt,
				Source:   "url:article",
				Metadata: article.Meta,
			}
			if doc.Metadata == nil {
				doc.Metadata = make(map[string]string)
			}
			// Ensure URL and title also appear in metadata.
			doc.Metadata["url"] = url
			if doc.Title != "" {
				doc.Metadata["title"] = doc.Title
			}
			res.PrimaryDocument = doc
			res.Notes = append(res.Notes, "HTML classified as article; Readability-style extraction succeeded.")
			return nil
		}
		res.Notes = append(res.Notes, "Article extraction either failed or produced empty content; falling back to structural HTML view.")
	}

	// Fallback: use structural HTML representation (headings/paragraphs).
	if parsedErr == nil && parsed != nil {
		// Build a synthetic content body from paragraphs.
		var contentBuilder strings.Builder
		for _, p := range parsed.Paragraphs {
			if strings.TrimSpace(p.Text) == "" {
				continue
			}
			if contentBuilder.Len() > 0 {
				contentBuilder.WriteString("\n\n")
			}
			contentBuilder.WriteString(p.Text)
		}
		content := strings.TrimSpace(contentBuilder.String())

		// Excerpt: use first paragraph or first 240 chars.
		excerpt := ""
		if len(parsed.Paragraphs) > 0 {
			excerpt = parsed.Paragraphs[0].Text
		}
		if len([]rune(excerpt)) > 240 {
			r := []rune(excerpt)
			excerpt = string(r[:240]) + "…"
		}

		doc := &SearchDocument{
			URL:      url,
			Kind:     SearchDocumentKindHTML,
			Title:    parsed.Title,
			Content:  content,
			HTML:     "", // we intentionally do not expose a full HTML fragment here
			Excerpt:  excerpt,
			Source:   "url:html",
			Metadata: parsed.Meta,
		}
		if doc.Metadata == nil {
			doc.Metadata = make(map[string]string)
		}
		doc.Metadata["url"] = url
		if doc.Title != "" {
			doc.Metadata["title"] = doc.Title
		}

		res.PrimaryDocument = doc
		res.Notes = append(res.Notes, "HTML processed as general page; using headings/paragraphs view.")
		return nil
	}

	// If parsing failed entirely, treat as binary/text fallback.
	res.Notes = append(res.Notes, "HTML parsing failed; treating content as binary.")
	return c.handleBinaryURL(url, idetect.TypeHTML, res)
}

// handleFeedURL processes an RSS/Atom feed retrieved from a URL.
func (c *Client) handleFeedURL(
	url string,
	body []byte,
	res *SearchResult,
) error {
	// First, ensure it actually looks like a feed.
	ft := irss.DetectFeedType(body)
	if ft == irss.FeedUnknown {
		res.Notes = append(res.Notes, "XML content did not look like RSS/Atom; not treated as feed.")
		// Fall back to treating as text.
		return c.handleTextURL(url, body, res)
	}

	feed, err := c.ParseRSS(body)
	if err != nil {
		res.Notes = append(res.Notes, "RSS/Atom parsing failed; falling back to text.")
		return c.handleTextURL(url, body, res)
	}

	res.Feed = feed

	// Summarize the feed into a SearchDocument.
	excerpt := ""
	if len(feed.Items) > 0 {
		first := feed.Items[0]
		if strings.TrimSpace(first.Description) != "" {
			excerpt = first.Description
		} else if strings.TrimSpace(first.Content) != "" {
			excerpt = first.Content
		} else {
			excerpt = first.Title
		}
		if len([]rune(excerpt)) > 240 {
			r := []rune(excerpt)
			excerpt = string(r[:240]) + "…"
		}
	}

	doc := &SearchDocument{
		URL:     url,
		Kind:    SearchDocumentKindFeed,
		Title:   feed.Title,
		Content: "", // feed content is represented primarily via Feed struct
		HTML:    "",
		Excerpt: excerpt,
		Source:  "url:rss",
		Metadata: map[string]string{
			"url":         url,
			"feed_link":   feed.Link,
			"feed_title":  feed.Title,
			"feed_items":  intToString(len(feed.Items)),
			"feed_source": "rss_or_atom",
		},
	}

	res.PrimaryDocument = doc
	res.Notes = append(res.Notes, "Content processed as RSS/Atom feed.")
	return nil
}

// handleJSONURL processes a JSON response as a generic JSON document.
//
// For now we treat JSON as opaque textual content; later stages may
// add specific detectors for schema-aware APIs (e.g. OpenAPI sources).
func (c *Client) handleJSONURL(
	url string,
	body []byte,
	res *SearchResult,
) error {
	text := strings.TrimSpace(string(body))
	// Truncate very long JSON for primary content.
	runes := []rune(text)
	if len(runes) > 4000 {
		text = string(runes[:4000]) + "…"
	}

	doc := &SearchDocument{
		URL:      url,
		Kind:     SearchDocumentKindJSON,
		Title:    "",
		Content:  text,
		HTML:     "",
		Excerpt:  makeExcerptFromText(text),
		Source:   "url:json",
		Metadata: map[string]string{"url": url},
	}
	res.PrimaryDocument = doc
	res.Notes = append(res.Notes, "Content processed as JSON document.")
	return nil
}

// handleTextURL processes plain-text content.
func (c *Client) handleTextURL(
	url string,
	body []byte,
	res *SearchResult,
) error {
	text := strings.TrimSpace(string(body))
	runes := []rune(text)
	if len(runes) > 4000 {
		text = string(runes[:4000]) + "…"
	}

	doc := &SearchDocument{
		URL:      url,
		Kind:     SearchDocumentKindText,
		Title:    "",
		Content:  text,
		HTML:     "",
		Excerpt:  makeExcerptFromText(text),
		Source:   "url:text",
		Metadata: map[string]string{"url": url},
	}
	res.PrimaryDocument = doc
	res.Notes = append(res.Notes, "Content processed as plain text.")
	return nil
}

// handleBinaryURL handles binary or unknown content types.
func (c *Client) handleBinaryURL(
	url string,
	rawType idetect.Type,
	res *SearchResult,
) error {
	doc := &SearchDocument{
		URL:      url,
		Kind:     SearchDocumentKindBinary,
		Title:    "",
		Content:  "",
		HTML:     "",
		Excerpt:  "",
		Source:   "url:binary",
		Metadata: map[string]string{"url": url, "raw_type": string(rawType)},
	}
	res.PrimaryDocument = doc
	res.Notes = append(res.Notes, "Content treated as binary or unsupported type.")
	return nil
}

// intToString converts an int to its string representation.
// A small helper to avoid importing strconv in multiple places.
func intToString(n int) string {
	// Simple, dependency-free conversion for small integers.
	// For correctness, we still use strconv internally.
	return strconvInt(n)
}

// strconvInt is a tiny wrapper to keep int→string conversion local.
func strconvInt(n int) string {
	return strconv.Itoa(n)
}

// makeExcerptFromText produces a short excerpt from arbitrary text.
func makeExcerptFromText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= 240 {
		return text
	}
	runes = runes[:240]
	s := string(runes)
	lastSpace := strings.LastIndex(s, " ")
	if lastSpace > 80 {
		s = s[:lastSpace]
	}
	return strings.TrimSpace(s) + "…"
}
