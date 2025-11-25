// aether/toon.go
//
// Public TOON (Token-Oriented Object Notation) API for Aether.
//
// This layer exposes:
//   • ToTOON(sr *SearchResult) *toon.Document
//   • MarshalTOON(sr *SearchResult) ([]byte, error)
//   • MarshalTOONPretty(sr *SearchResult) ([]byte, error)
//
// Internal pipeline:
//   1. aether.NormalizeSearchResult() → model.Document
//   2. internal/toon.FromModel()      → toon.Document
//
// The TOON representation is stable, deterministic, and optimized
// for LLM consumption, structured reasoning, and interop with tools.

package aether

import (
	"encoding/json"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

//
// ─────────────────────────────────────────────────────────────────────────────
//                           CORE CONVERSION (PUBLIC)
// ─────────────────────────────────────────────────────────────────────────────
//

// ToTOON converts a public SearchResult into a TOON 2.0 Document.
//
// Steps:
//  1. NormalizeSearchResult → model.Document
//  2. toon.FromModel        → toon.Document
func (c *Client) ToTOON(sr *SearchResult) *toon.Document {
	if c == nil || sr == nil {
		return &toon.Document{}
	}

	// Convert SearchResult → normalized model.Document
	normalized := c.NormalizeSearchResult(sr)
	if normalized == nil {
		return &toon.Document{}
	}

	// Convert model.Document → TOON Document
	return toon.FromModel(normalized)
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                             JSON SERIALIZATION
// ─────────────────────────────────────────────────────────────────────────────
//

// MarshalTOON serializes a SearchResult into compact TOON JSON.
func (c *Client) MarshalTOON(sr *SearchResult) ([]byte, error) {
	doc := c.ToTOON(sr)
	return json.Marshal(doc)
}

// MarshalTOONPretty serializes a SearchResult into pretty-printed TOON JSON.
func (c *Client) MarshalTOONPretty(sr *SearchResult) ([]byte, error) {
	doc := c.ToTOON(sr)
	return json.MarshalIndent(doc, "", "  ")
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                        DIRECT MODEL → TOON HELPERS
// ─────────────────────────────────────────────────────────────────────────────
//

// ToTOONFromModel converts an internal normalized document directly to TOON.
// Useful for embedding Aether as a library when skipping SearchResult pipeline.
func (c *Client) ToTOONFromModel(doc *model.Document) *toon.Document {
	if doc == nil {
		return &toon.Document{}
	}
	return toon.FromModel(doc)
}

// MarshalTOONFromModel serializes a normalized Document into compact TOON JSON.
func (c *Client) MarshalTOONFromModel(doc *model.Document) ([]byte, error) {
	tdoc := c.ToTOONFromModel(doc)
	return json.Marshal(tdoc)
}

// MarshalTOONPrettyFromModel serializes a normalized Document into pretty JSON.
func (c *Client) MarshalTOONPrettyFromModel(doc *model.Document) ([]byte, error) {
	tdoc := c.ToTOONFromModel(doc)
	return json.MarshalIndent(tdoc, "", "  ")
}
