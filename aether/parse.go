// aether/parse.go
//
// This file exposes a high-level HTML parsing API on the Aether Client.
// It builds on the internal/html package to provide a stable, public
// representation of headings, paragraphs, links and metadata.
//
// Aether.ParseHTML operates purely on HTML bytes; fetching and robots.txt
// handling are provided separately by Client.Fetch.

package aether

import (
	"fmt"

	ihtml "github.com/Nibir1/Aether/internal/html"
)

// ParsedHTML represents a normalized view of an HTML document.
// It is designed to be LLM-friendly and stable as a public API.
type ParsedHTML struct {
	Title      string
	Headings   []Heading
	Paragraphs []Paragraph
	Links      []Link
	Meta       map[string]string
}

// Heading represents a heading in the document (h1–h6).
type Heading struct {
	Level int
	Text  string
}

// Paragraph represents paragraph text in the document.
type Paragraph struct {
	Text string
}

// Link represents a hyperlink in the document.
type Link struct {
	Href string
	Text string
	Rel  string
}

// ParseHTML parses raw HTML bytes into a ParsedHTML structure.
//
// This method does not perform any network operations and does not
// consult robots.txt. It is intended for HTML you already have, such
// as the body of a FetchResult.
func (c *Client) ParseHTML(html []byte) (*ParsedHTML, error) {
	if c == nil {
		return nil, fmt.Errorf("aether: nil client in ParseHTML")
	}

	doc, err := ihtml.ParseDocument(html)
	if err != nil {
		return nil, err
	}

	title := ihtml.ExtractTitle(doc)
	meta := ihtml.ExtractMeta(doc)
	headingsInternal := ihtml.ExtractHeadings(doc)
	paragraphsInternal := ihtml.ExtractParagraphs(doc)
	linksInternal := ihtml.ExtractLinks(doc)

	headings := make([]Heading, 0, len(headingsInternal))
	for _, h := range headingsInternal {
		headings = append(headings, Heading{
			Level: h.Level,
			Text:  h.Text,
		})
	}

	paragraphs := make([]Paragraph, 0, len(paragraphsInternal))
	for _, p := range paragraphsInternal {
		paragraphs = append(paragraphs, Paragraph{
			Text: p.Text,
		})
	}

	links := make([]Link, 0, len(linksInternal))
	for _, l := range linksInternal {
		links = append(links, Link{
			Href: l.Href,
			Text: l.Text,
			Rel:  l.Rel,
		})
	}

	return &ParsedHTML{
		Title:      title,
		Headings:   headings,
		Paragraphs: paragraphs,
		Links:      links,
		Meta:       meta,
	}, nil
}
