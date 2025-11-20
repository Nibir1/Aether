// internal/extract/readability.go
//
// Package extract implements Aether's article extraction engine.
// It performs Readability-style scoring on an HTML document to identify
// the main content block, returning a normalized Article structure.
//
// The goal is not to perfectly replicate Mozilla Readability, but to
// provide a robust, deterministic, and explainable approximation that
// works well on most news and blog content.

package extract

import (
	"bytes"
	"strings"

	ihtml "github.com/Nibir1/Aether/internal/html"
	xhtml "golang.org/x/net/html"
)

// Article represents the extracted main content of an HTML document.
//
// ContentHTML contains a sanitized HTML fragment of the main article.
// Text contains plain text derived from ContentHTML.
// Excerpt is a short summary derived from the beginning of the Text.
type Article struct {
	Title       string
	Byline      string
	ContentHTML string
	Text        string
	Excerpt     string
	TopImageURL string
}

// Extract runs the Readability-style algorithm on a parsed HTML Document.
//
// baseURL is optional but may be used in future enhancements to resolve
// relative URLs or canonical links.
func Extract(doc *ihtml.Document, baseURL string) *Article {
	if doc == nil || doc.Root == nil {
		return &Article{}
	}

	body := findBodyNode(doc.Root)
	if body == nil {
		body = doc.Root
	}

	// Clean the DOM: ignore obvious boilerplate tags (nav, aside, footer, etc.)
	cleanNodeTree(body)

	// Score candidate nodes and pick the best container for the main content.
	candidates := scoreCandidates(body)
	top := selectTopCandidate(candidates)
	if top == nil {
		// Fallback: use entire body text if no candidate is found.
		text := nodeText(body)
		text = strings.TrimSpace(text)
		return &Article{
			Title:   "",
			Text:    text,
			Excerpt: makeExcerpt(text),
		}
	}

	// Build a content fragment around the top candidate and its siblings.
	contentNode := buildContentNode(top)

	var buf bytes.Buffer
	if err := xhtml.Render(&buf, contentNode); err != nil {
		// On rendering failure, fallback to text-only extraction.
		txt := nodeText(contentNode)
		txt = strings.TrimSpace(txt)
		return &Article{
			Title:   "",
			Text:    txt,
			Excerpt: makeExcerpt(txt),
		}
	}

	html := buf.String()
	text := nodeText(contentNode)
	text = strings.TrimSpace(text)

	return &Article{
		Title:       "",
		ContentHTML: html,
		Text:        text,
		Excerpt:     makeExcerpt(text),
	}
}
