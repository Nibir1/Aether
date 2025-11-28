// aether/toon.go
//
// Public TOON (Token-Oriented Object Notation) API for Aether.
//
// This layer exposes conversion functions from Aether’s high-level
// SearchResult and normalized model.Document into TOON 2.0 structures.
//
//   • ToTOON(sr *SearchResult) *toon.Document
//   • MarshalTOON(sr *SearchResult) ([]byte, error)
//   • MarshalTOONPretty(sr *SearchResult) ([]byte, error)
//
// Pipeline:
//   1. NormalizeSearchResult() → *model.Document
//   2. toon.FromModel()        → *toon.Document
//
// The TOON representation is stable and structured for LLM consumption.

package aether

import (
	"encoding/json"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

//
// ─────────────────────────────────────────────────────────────────────────────
//                               CORE CONVERSION
// ─────────────────────────────────────────────────────────────────────────────
//

// ToTOON converts a SearchResult into a TOON 2.0 document.
// Safe for nil receivers and nil SearchResult.
func (c *Client) ToTOON(sr *SearchResult) *toon.Document {
	if c == nil || sr == nil {
		return &toon.Document{}
	}

	normalized := c.NormalizeSearchResult(sr)
	if normalized == nil {
		return &toon.Document{}
	}

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

// MarshalTOONPretty serializes a SearchResult into pretty-printed JSON.
func (c *Client) MarshalTOONPretty(sr *SearchResult) ([]byte, error) {
	doc := c.ToTOON(sr)
	return json.MarshalIndent(doc, "", "  ")
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                       DIRECT MODEL → TOON CONVERSIONS
// ─────────────────────────────────────────────────────────────────────────────
//

// ToTOONFromModel converts a normalized internal model.Document directly
// into a TOON document. Useful when embedding Aether as a library
// without using SearchResult.
func (c *Client) ToTOONFromModel(doc *model.Document) *toon.Document {
	if doc == nil {
		return &toon.Document{}
	}
	return toon.FromModel(doc)
}

// MarshalTOONFromModel serializes a normalized model.Document into compact JSON.
func (c *Client) MarshalTOONFromModel(doc *model.Document) ([]byte, error) {
	tdoc := c.ToTOONFromModel(doc)
	return json.Marshal(tdoc)
}

// MarshalTOONPrettyFromModel serializes a normalized model.Document into pretty JSON.
func (c *Client) MarshalTOONPrettyFromModel(doc *model.Document) ([]byte, error) {
	tdoc := c.ToTOONFromModel(doc)
	return json.MarshalIndent(tdoc, "", "  ")
}
