// internal/toon/types.go
//
// Core TOON 2.0 schema — the stable, LLM-friendly token representation.
// This file defines the fundamental types used throughout Aether’s TOON
// system: TokenType, Token, and Document.
//
// TOON (Token-Oriented Object Notation) is designed for:
//   • LLM-friendly, structure-preserving serialization
//   • deterministic document structure
//   • lossless representation of Aether’s normalized model.Document
//   • plugin-extendable metadata via TokenMeta and Document.Attributes
//
// The token stream expresses a normalized document as:
//
//   DOCINFO
//   TITLE?
//   EXCERPT?
//   (TEXT)?                // if no sections
//   SECTION_START+...
//       HEADING?
//       TEXT*
//       META*
//   SECTION_END+...
//
// This structure is stable and version-safe.

package toon

import "github.com/Nibir1/Aether/internal/model"

//
// ───────────────────────────────────────────────────────────────
//                           TOKEN TYPES
// ───────────────────────────────────────────────────────────────
//

// TokenType represents the atomic category of a TOON token.
type TokenType string

const (
	// Plain text content (paragraphs)
	TokenText TokenType = "text"

	// Heading inside a section (NOT the document title)
	TokenHeading TokenType = "heading"

	// Section enter/leave
	TokenSectionStart TokenType = "section_start"
	TokenSectionEnd   TokenType = "section_end"

	// Metadata key/value pairs
	TokenMeta TokenType = "meta"

	// Document-level info (kind, high-level attrs)
	TokenDocumentInfo TokenType = "docinfo"

	// NEW in TOON 2.0:
	// Dedicated top-level title + excerpt tokens
	TokenTitle   TokenType = "title"
	TokenExcerpt TokenType = "excerpt"
)

//
// ───────────────────────────────────────────────────────────────
//                              TOKEN
// ───────────────────────────────────────────────────────────────
//

// Token is the atomic TOON unit. It may represent:
//
//   - raw text
//   - heading
//   - structural boundary
//   - metadata key/value
//   - document identity
//   - title/excerpt metadata
//
// Role is optional — content tokens inside sections use the section's
// role (e.g. "body", "feed_item", "entity") for semantic grouping.
type Token struct {
	Type  TokenType         `json:"type"`
	Role  string            `json:"role,omitempty"`  // semantic role (body, entity, feed_item)
	Text  string            `json:"text,omitempty"`  // text for text/heading/title/excerpt
	Attrs map[string]string `json:"attrs,omitempty"` // arbitrary metadata
}

//
// ───────────────────────────────────────────────────────────────
//                           TOON DOCUMENT
// ───────────────────────────────────────────────────────────────
//

// Document is the full TOON representation of a normalized document.
// It contains:
//
//   - High-level identity (URL, kind)
//   - Optional top-level title/excerpt
//   - Full ordered token stream
//   - Extra attributes (flattened metadata)
//
// This is intentionally decoupled from model.Document so TOON remains
// stable even if model.Document evolves.
type Document struct {
	// Top-level identity
	SourceURL string             `json:"source_url,omitempty"`
	Kind      model.DocumentKind `json:"kind"`

	// High-level normalized metadata
	Title   string `json:"title,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`

	// Ordered token stream (TOON core)
	Tokens []Token `json:"tokens,omitempty"`

	// Extra flattened metadata (model.Document.Metadata)
	Attributes map[string]string `json:"attributes,omitempty"`
}
