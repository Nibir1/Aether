// aether/search.go
//
// High-level search pipeline for Aether.
//
// Aether.Search is the primary entrypoint for turning an arbitrary user
// query into a structured SearchResult that can then be normalized into
// JSON or TOON and rendered for LLM consumption.
//
// Responsibilities:
//   • Classify query (URL vs free-text lookup)
//   • Route to SourcePlugins where available
//   • Fallback to built-in OpenAPI integrations (e.g. Wikipedia)
//   • Perform direct HTTP fetch for URL queries
//   • Produce a SearchResult with a PrimaryDocument, optional Article/Feed,
//     and a SearchPlan describing what was done.
//
// NOTE:
// This is the first full implementation of the search pipeline. Future
// stages may extend it with:
//   • richer SmartQuery intent detection
//   • multi-source federation and merging
//   • deeper RSS/article handling
//   • plugin transform pipelines.

package aether

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nibir1/Aether/plugins"
)

//
// ────────────────────────────────────────────────
//                  SEARCH TYPES
// ────────────────────────────────────────────────
//

// SearchIntent describes the broad category of a query.
type SearchIntent string

const (
	SearchIntentUnknown SearchIntent = "unknown"
	SearchIntentURL     SearchIntent = "url"
	SearchIntentLookup  SearchIntent = "lookup"
	SearchIntentPlugin  SearchIntent = "plugin"
)

// SearchPlan describes how Aether decided to handle a query.
type SearchPlan struct {
	RawQuery string       // original user query
	Intent   SearchIntent // URL vs lookup vs plugin-driven
	URL      string       // populated for URL-based queries
	Source   string       // plugin or integration that provided primary result
}

// SearchDocumentKind describes the kind of the primary document.
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

// SearchDocument is the primary, high-level document for a search.
//
// It intentionally mirrors Aether's normalized model at a higher level,
// and is later converted into model.Document by NormalizeSearchResult.
type SearchDocument struct {
	URL      string             // canonical URL, if any
	Kind     SearchDocumentKind // article, html_page, feed, json, text, binary
	Title    string             // human-readable title
	Excerpt  string             // short summary or snippet
	Content  string             // main textual body
	Metadata map[string]string  // flat key/value metadata
}

// SearchResult is the main output of Aether.Search.
//
// It contains:
//   - the original query
//   - the plan describing what was done
//   - the primary document (SearchDocument)
//   - optional richer views (Article, Feed) populated by other subsystems
type SearchResult struct {
	Query           string
	Plan            SearchPlan
	PrimaryDocument *SearchDocument

	// Optional views that may be populated by other subsystems:
	Article *Article // from ExtractText / readability engine (Stage 5)
	Feed    *Feed    // from RSS subsystem (Stage 8)
}

//
// ────────────────────────────────────────────────
//                    ENTRYPOINT
// ────────────────────────────────────────────────
//

// Search runs Aether's high-level search pipeline for the given query.
//
// Behavior:
//   - URL queries → robots.txt-compliant fetch → SearchDocument
//   - Non-URL queries:
//   - Try SourcePlugins (in registration order)
//   - Fallback to Wikipedia summary via internal OpenAPI client
//
// TODO (future stages): expand SmartQuery classification, integrate RSS
// auto-detection, transform plugins, and richer federation.
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error) {
	if c == nil {
		return nil, fmt.Errorf("aether: nil client in Search")
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("aether: empty query")
	}

	plan := SearchPlan{
		RawQuery: query,
		Intent:   SearchIntentUnknown,
	}

	if isProbablyURL(query) {
		plan.Intent = SearchIntentURL
		plan.URL = query

		doc, err := c.searchURL(ctx, plan)
		if err != nil {
			return nil, err
		}

		return &SearchResult{
			Query:           query,
			Plan:            plan,
			PrimaryDocument: doc,
			Article:         nil,
			Feed:            nil,
		}, nil
	}

	// non-URL → treat as lookup / textual query
	plan.Intent = SearchIntentLookup

	// 1) Try SourcePlugins first (plugin-based integrations).
	if c.plugins != nil {
		if doc, sourceName, err := c.searchViaPlugins(ctx, query); err == nil && doc != nil {
			plan.Intent = SearchIntentPlugin
			plan.Source = sourceName

			return &SearchResult{
				Query:           query,
				Plan:            plan,
				PrimaryDocument: doc,
				Article:         nil,
				Feed:            nil,
			}, nil
		}
	}

	// 2) Fallback to Wikipedia summary for factual queries.
	doc, err := c.searchViaWikipedia(ctx, query)
	if err != nil {
		return nil, err
	}
	plan.Source = "wikipedia"

	return &SearchResult{
		Query:           query,
		Plan:            plan,
		PrimaryDocument: doc,
		Article:         nil,
		Feed:            nil,
	}, nil
}

//
// ────────────────────────────────────────────────
//                 URL-BASED SEARCH
// ────────────────────────────────────────────────
//

func (c *Client) searchURL(ctx context.Context, plan SearchPlan) (*SearchDocument, error) {
	body, headers, err := c.FetchRaw(ctx, plan.URL)
	if err != nil {
		return nil, err
	}

	contentType := classifyContentType(headers)
	textBody := string(body)

	kind := SearchDocumentKindUnknown
	switch {
	case strings.Contains(contentType, "html"):
		kind = SearchDocumentKindHTML
	case strings.Contains(contentType, "json"):
		kind = SearchDocumentKindJSON
	case strings.HasPrefix(contentType, "text/"):
		kind = SearchDocumentKindText
	default:
		kind = SearchDocumentKindBinary
	}

	metadata := map[string]string{
		"content_type": contentType,
		"source":       "direct_fetch",
	}

	excerpt := buildExcerpt(textBody, 320)

	return &SearchDocument{
		URL:      plan.URL,
		Kind:     kind,
		Title:    "", // Article extractor / metadata subsystem may fill this later.
		Excerpt:  excerpt,
		Content:  textBody,
		Metadata: metadata,
	}, nil
}

//
// ────────────────────────────────────────────────
//                 PLUGIN-BASED SEARCH
// ────────────────────────────────────────────────
//

// searchViaPlugins attempts to satisfy the query using registered SourcePlugins.
// It returns the first successful SearchDocument produced by any plugin.
func (c *Client) searchViaPlugins(ctx context.Context, query string) (*SearchDocument, string, error) {
	if c.plugins == nil {
		return nil, "", fmt.Errorf("no plugin registry available")
	}

	names := c.plugins.ListSources()
	for _, name := range names {
		p := c.plugins.GetSource(name)
		if p == nil {
			continue
		}

		doc, err := p.Fetch(ctx, query)
		if err != nil || doc == nil {
			continue
		}

		sd := searchDocumentFromPluginDocument(doc)
		if sd == nil {
			continue
		}

		// Attach plugin name into metadata
		if sd.Metadata == nil {
			sd.Metadata = map[string]string{}
		}
		sd.Metadata["aether.source_plugin"] = name

		return sd, name, nil
	}

	return nil, "", fmt.Errorf("no source plugin produced a result")
}

// searchDocumentFromPluginDocument converts a plugins.Document into a
// SearchDocument suitable for Aether's SearchResult.
func searchDocumentFromPluginDocument(doc *plugins.Document) *SearchDocument {
	if doc == nil {
		return nil
	}

	kind := SearchDocumentKindUnknown
	switch doc.Kind {
	case plugins.DocumentKindArticle:
		kind = SearchDocumentKindArticle
	case plugins.DocumentKindHTML:
		kind = SearchDocumentKindHTML
	case plugins.DocumentKindFeed:
		kind = SearchDocumentKindFeed
	case plugins.DocumentKindJSON:
		kind = SearchDocumentKindJSON
	case plugins.DocumentKindText:
		kind = SearchDocumentKindText
	case plugins.DocumentKindBinary:
		kind = SearchDocumentKindBinary
	default:
		kind = SearchDocumentKindUnknown
	}

	meta := map[string]string{}
	for k, v := range doc.Metadata {
		meta[k] = v
	}
	if doc.Source != "" {
		meta["aether.plugin_source"] = doc.Source
	}

	excerpt := doc.Excerpt
	if strings.TrimSpace(excerpt) == "" {
		excerpt = buildExcerpt(doc.Content, 320)
	}

	return &SearchDocument{
		URL:      doc.URL,
		Kind:     kind,
		Title:    doc.Title,
		Excerpt:  excerpt,
		Content:  doc.Content,
		Metadata: meta,
	}
}

//
// ────────────────────────────────────────────────
//              WIKIPEDIA FALLBACK SEARCH
// ────────────────────────────────────────────────
//

// searchViaWikipedia uses the built-in OpenAPI client to fetch a concise
// summary of the topic from Wikipedia.
//
// This is the default fallback for non-URL queries when no SourcePlugin
// handled the request.
func (c *Client) searchViaWikipedia(ctx context.Context, query string) (*SearchDocument, error) {
	summary, err := c.WikipediaSummary(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("aether: wikipedia lookup failed: %w", err)
	}
	if summary == nil {
		return nil, fmt.Errorf("aether: wikipedia returned no data")
	}

	meta := map[string]string{
		"source":   "wikipedia",
		"lang":     summary.Language,
		"page_url": summary.URL,
	}

	excerpt := summary.Description
	if strings.TrimSpace(excerpt) == "" {
		excerpt = buildExcerpt(summary.Extract, 320)
	}

	return &SearchDocument{
		URL:      summary.URL,
		Kind:     SearchDocumentKindArticle,
		Title:    summary.Title,
		Excerpt:  excerpt,
		Content:  summary.Extract,
		Metadata: meta,
	}, nil
}

//
// ────────────────────────────────────────────────
//                  HELPER FUNCTIONS
// ────────────────────────────────────────────────
//

// isProbablyURL performs a lightweight heuristic to decide whether a query
// looks like a URL (http/https).
func isProbablyURL(q string) bool {
	if strings.HasPrefix(q, "http://") || strings.HasPrefix(q, "https://") {
		u, err := url.Parse(q)
		return err == nil && u.Scheme != "" && u.Host != ""
	}
	return false
}

// classifyContentType normalizes a Content-Type header into a simple string.
func classifyContentType(h http.Header) string {
	if h == nil {
		return "application/octet-stream"
	}
	ct := h.Get("Content-Type")
	if ct == "" {
		return "application/octet-stream"
	}
	return strings.ToLower(ct)
}

// buildExcerpt produces a short snippet from a longer body.
func buildExcerpt(body string, maxLen int) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	// Collapse whitespace and newlines for a single-line excerpt.
	body = collapseWhitespace(body)

	if len(body) <= maxLen {
		return body
	}
	if maxLen <= 3 {
		return body[:maxLen]
	}

	return body[:maxLen-3] + "..."
}

// collapseWhitespace replaces all sequences of whitespace with a single space.
func collapseWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	space := false

	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			if !space {
				b.WriteRune(' ')
				space = true
			}
		} else {
			b.WriteRune(r)
			space = false
		}
	}
	return strings.TrimSpace(b.String())
}
