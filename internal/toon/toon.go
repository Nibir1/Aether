// internal/toon/toon.go
//
// Package toon defines Aether's TOON (Token-Oriented Object Notation)
// representation and provides a converter from the internal model.Document.
//
// TOON is still JSON at the transport level, but it is organized as a
// sequence of tokens with explicit semantic roles. This makes it easier
// for LLMs to consume and reason about content without guessing structure.

package toon

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// TokenType describes the type of a TOON token.
type TokenType string

const (
	TokenTypeField     TokenType = "field"     // key/value field (title, url, etc.)
	TokenTypeParagraph TokenType = "paragraph" // body paragraph
	TokenTypeMeta      TokenType = "meta"      // metadata field
	TokenTypeSection   TokenType = "section"   // section marker
)

// Token represents a single semantic unit in TOON.
type Token struct {
	Type TokenType `json:"type"`
	Role string    `json:"role,omitempty"` // e.g. "title", "body", "summary", "feed_item"
	Key  string    `json:"key,omitempty"`
	Text string    `json:"text,omitempty"`
}

// Document is the TOON representation of a normalized document.
type Document struct {
	Schema string            `json:"schema"`         // e.g. "aether.toon.v1"
	Kind   string            `json:"kind"`           // mirrors model.Document.Kind
	URL    string            `json:"url,omitempty"`  // source URL
	Meta   map[string]string `json:"meta,omitempty"` // high-level metadata
	Tokens []Token           `json:"tokens"`         // token sequence
}

// FromModel converts a normalized model.Document into a TOON Document.
func FromModel(doc *model.Document) *Document {
	if doc == nil {
		return &Document{
			Schema: "aether.toon.v1",
			Kind:   string(model.DocumentKindUnknown),
			URL:    "",
			Meta:   map[string]string{},
			Tokens: nil,
		}
	}

	out := &Document{
		Schema: "aether.toon.v1",
		Kind:   string(doc.Kind),
		URL:    doc.SourceURL,
		Meta:   map[string]string{},
		Tokens: make([]Token, 0, 32),
	}

	// Copy metadata into Meta and also produce meta tokens.
	for k, v := range doc.Metadata {
		if out.Meta == nil {
			out.Meta = map[string]string{}
		}
		out.Meta[k] = v
		out.Tokens = append(out.Tokens, Token{
			Type: TokenTypeMeta,
			Role: "meta",
			Key:  k,
			Text: v,
		})
	}

	// Title field token.
	if strings.TrimSpace(doc.Title) != "" {
		out.Tokens = append(out.Tokens, Token{
			Type: TokenTypeField,
			Role: "title",
			Key:  "title",
			Text: doc.Title,
		})
	}

	// Excerpt field token.
	if strings.TrimSpace(doc.Excerpt) != "" {
		out.Tokens = append(out.Tokens, Token{
			Type: TokenTypeField,
			Role: "summary",
			Key:  "excerpt",
			Text: doc.Excerpt,
		})
	}

	// Content paragraphs: split on blank lines to produce paragraph tokens.
	if strings.TrimSpace(doc.Content) != "" {
		paragraphs := splitParagraphs(doc.Content)
		for _, p := range paragraphs {
			if strings.TrimSpace(p) == "" {
				continue
			}
			out.Tokens = append(out.Tokens, Token{
				Type: TokenTypeParagraph,
				Role: "body",
				Text: p,
			})
		}
	}

	// Sections: section markers + paragraph tokens.
	for _, s := range doc.Sections {
		role := string(s.Role)
		if role == "" {
			role = string(model.SectionRoleUnknown)
		}

		// Section marker
		out.Tokens = append(out.Tokens, Token{
			Type: TokenTypeSection,
			Role: role,
			Key:  s.Heading,
			Text: "",
		})

		// Section text as paragraphs.
		if strings.TrimSpace(s.Text) != "" {
			paragraphs := splitParagraphs(s.Text)
			for _, p := range paragraphs {
				if strings.TrimSpace(p) == "" {
					continue
				}
				out.Tokens = append(out.Tokens, Token{
					Type: TokenTypeParagraph,
					Role: role,
					Text: p,
				})
			}
		}

		// Section-level meta turned into meta tokens.
		for k, v := range s.Meta {
			out.Tokens = append(out.Tokens, Token{
				Type: TokenTypeMeta,
				Role: role,
				Key:  k,
				Text: v,
			})
		}
	}

	return out
}

// splitParagraphs splits text into paragraphs on double newlines,
// trimming whitespace around each paragraph.
func splitParagraphs(text string) []string {
	chunks := strings.Split(text, "\n\n")
	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		t := strings.TrimSpace(c)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
