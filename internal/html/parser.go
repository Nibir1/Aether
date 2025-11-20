// internal/html/parser.go
//
// Package html implements Aether's internal HTML parsing utilities.
// It wraps golang.org/x/net/html to provide a higher-level Document
// abstraction and helpers for extracting headings, paragraphs, links,
// and metadata.
//
// This package is internal because Aether's public API exposes its
// own stable types and does not leak third-party representations.

package html

import (
	"bytes"

	xhtml "golang.org/x/net/html"
)

// Document represents a parsed HTML document.
//
// Root is the root node returned by the underlying HTML parser.
type Document struct {
	Root *xhtml.Node
}

// ParseDocument parses raw HTML bytes into a Document.
//
// It uses golang.org/x/net/html for robust HTML5 parsing.
func ParseDocument(data []byte) (*Document, error) {
	root, err := xhtml.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &Document{Root: root}, nil
}
