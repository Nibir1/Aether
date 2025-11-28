// aether/toon_bton.go
//
// Public BTON (Binary TOON) serializers for Aether.
//
// This file exposes stable wrappers around the internal toon.EncodeBTON /
// DecodeBTON implementations, providing:
//
//   • MarshalBTON(sr *SearchResult) → []byte
//   • UnmarshalBTON(data []byte)    → *toon.Document
//   • MarshalBTONFromModel(doc)     → []byte
//
// All methods are safe for nil receivers and nil input values. Returned
// documents always reference TOON 2.0 schema.
//

package aether

import (
	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

//
// ─────────────────────────────────────────────────────────────────────────────
//                         SEARCHRESULT → BTON
// ─────────────────────────────────────────────────────────────────────────────
//

// MarshalBTON serializes a SearchResult into BT0N binary format.
// Nil client or nil SearchResult returns an empty TOON document.
func (c *Client) MarshalBTON(sr *SearchResult) ([]byte, error) {
	if c == nil || sr == nil {
		empty := &toon.Document{}
		return toon.EncodeBTON(empty)
	}
	tdoc := c.ToTOON(sr)
	return toon.EncodeBTON(tdoc)
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                               BTON → TOON
// ─────────────────────────────────────────────────────────────────────────────
//

// UnmarshalBTON parses BT0N bytes into a TOON Document.
// If data is nil or empty, DecodeBTON handles it gracefully.
func (c *Client) UnmarshalBTON(data []byte) (*toon.Document, error) {
	return toon.DecodeBTON(data)
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                          MODEL DOCUMENT → BTON
// ─────────────────────────────────────────────────────────────────────────────
//

// MarshalBTONFromModel serializes a normalized model.Document into BT0N.
func (c *Client) MarshalBTONFromModel(doc *model.Document) ([]byte, error) {
	if c == nil || doc == nil {
		empty := &toon.Document{}
		return toon.EncodeBTON(empty)
	}
	tdoc := c.ToTOONFromModel(doc)
	return toon.EncodeBTON(tdoc)
}
