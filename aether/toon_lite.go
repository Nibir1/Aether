// aether/toon_lite.go
//
// Public TOON-Lite JSON serializers for Aether.
//
// This layer exposes compact JSON ("TOON-Lite") versions of TOON documents,
// optimized for:
//    • vector databases
//    • embedding pipelines
//    • compact storage
//
// API surface:
//    - MarshalTOONLite(sr *SearchResult) ([]byte, error)
//    - MarshalTOONLitePretty(sr *SearchResult) ([]byte, error)
//    - MarshalTOONLiteFromModel(doc *model.Document) ([]byte, error)
//    - MarshalTOONLitePrettyFromModel(doc *model.Document) ([]byte, error)
//

package aether

import (
	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

//
// ─────────────────────────────────────────────────────────────
//          SEARCHRESULT → TOON-LITE JSON
// ─────────────────────────────────────────────────────────────
//

// MarshalTOONLite serializes SearchResult → normalized → TOON → Lite JSON.
func (c *Client) MarshalTOONLite(sr *SearchResult) ([]byte, error) {
	if c == nil || sr == nil {
		empty := &toon.Document{}
		return toon.MarshalLite(empty)
	}
	tdoc := c.ToTOON(sr)
	return toon.MarshalLite(tdoc)
}

// MarshalTOONLitePretty serializes SearchResult into pretty-printed Lite JSON.
func (c *Client) MarshalTOONLitePretty(sr *SearchResult) ([]byte, error) {
	if c == nil || sr == nil {
		empty := &toon.Document{}
		return toon.MarshalLitePretty(empty)
	}
	tdoc := c.ToTOON(sr)
	return toon.MarshalLitePretty(tdoc)
}

//
// ─────────────────────────────────────────────────────────────
//          MODEL DOCUMENT → TOON-LITE JSON
// ─────────────────────────────────────────────────────────────
//

// MarshalTOONLiteFromModel serializes a normalized model.Document → Lite JSON.
func (c *Client) MarshalTOONLiteFromModel(doc *model.Document) ([]byte, error) {
	if c == nil || doc == nil {
		empty := &toon.Document{}
		return toon.MarshalLite(empty)
	}
	tdoc := c.ToTOONFromModel(doc)
	return toon.MarshalLite(tdoc)
}

// MarshalTOONLitePrettyFromModel pretty-prints model.Document → Lite JSON.
func (c *Client) MarshalTOONLitePrettyFromModel(doc *model.Document) ([]byte, error) {
	if c == nil || doc == nil {
		empty := &toon.Document{}
		return toon.MarshalLitePretty(empty)
	}
	tdoc := c.ToTOONFromModel(doc)
	return toon.MarshalLitePretty(tdoc)
}
