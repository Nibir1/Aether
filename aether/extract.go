// aether/extract.go
//
// This file exposes Aether's article extraction interface. It builds on
// the internal HTML parser and Readability-style extractor to convert
// arbitrary HTML documents (or fetched URLs) into normalized Article
// structures that are friendly for LLM consumption.

package aether

import (
	"context"
	"fmt"

	iextract "github.com/Nibir1/Aether/internal/extract"
	ihtml "github.com/Nibir1/Aether/internal/html"
)

//
// ───────────────────────────────────────────────────────────────
//                           PUBLIC TYPES
// ───────────────────────────────────────────────────────────────
//

// Article is the public article representation returned by Aether.
//
// Content is the plain-text main content.
// HTML is a sanitized HTML fragment of the main content.
// Excerpt is a short summary derived from the article body.
// Meta contains document metadata extracted from <meta> tags.
type Article struct {
	URL     string
	Title   string
	Byline  string
	Content string
	HTML    string
	Excerpt string
	Meta    map[string]string
}

//
// ───────────────────────────────────────────────────────────────
//                  EXTRACT FROM RAW HTML (NO FETCH)
// ───────────────────────────────────────────────────────────────
//

// ExtractArticleFromHTML extracts the main article content from raw HTML.
//
// url is optional but recommended; it is stored in the Article result
// and may be used by future features (e.g. canonical URL resolution).
func (c *Client) ExtractArticleFromHTML(html []byte, url string) (*Article, error) {
	if len(html) == 0 {
		return nil, fmt.Errorf("aether: empty HTML buffer")
	}

	doc, err := ihtml.ParseDocument(html)
	if err != nil {
		return nil, err
	}

	// Base HTML metadata
	title := ihtml.ExtractTitle(doc)
	meta := ihtml.ExtractMeta(doc)

	// Run Readability-style extraction
	internal := iextract.Extract(doc, url)
	if internal == nil {
		internal = &iextract.Article{}
	}

	// Prefer extractor title if present; otherwise use <title>.
	finalTitle := internal.Title
	if finalTitle == "" {
		finalTitle = title
	}

	article := &Article{
		URL:     url,
		Title:   finalTitle,
		Byline:  internal.Byline,
		Content: internal.Text,
		HTML:    internal.ContentHTML,
		Excerpt: internal.Excerpt,
		Meta:    meta,
	}
	return article, nil
}

//
// ───────────────────────────────────────────────────────────────
//                  HIGH-LEVEL: FETCH + EXTRACT
// ───────────────────────────────────────────────────────────────
//

// ExtractArticle fetches the given URL (respecting robots.txt) and runs
// article extraction on the retrieved HTML.
//
// This is a convenience wrapper around Fetch + ExtractArticleFromHTML.
func (c *Client) ExtractArticle(ctx context.Context, url string) (*Article, error) {
	if c == nil {
		return nil, fmt.Errorf("aether: nil client in ExtractArticle")
	}
	if url == "" {
		return nil, fmt.Errorf("aether: empty URL in ExtractArticle")
	}

	res, err := c.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	return c.ExtractArticleFromHTML(res.Body, url)
}
