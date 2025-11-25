// internal/toon/grammar.go
//
// Grammar helpers and classification functions for TOON tokens.
//
// This file does not redefine core types; it adds semantic helpers on
// top of TokenType and Token, describing how the token stream is
// interpreted.
//
// Grammar (high-level):
//
//   Document:
//     - optional DOCINFO token
//     - optional TITLE token
//     - optional EXCERPT token
//     - zero or more SECTION blocks
//
//   SECTION block:
//     - SECTION_START (with role, optional heading)
//     - optional HEADING token
//     - zero or more TEXT / META tokens
//     - SECTION_END
//
//   Additional tokens:
//     - META tokens may appear at document level or inside sections.

package toon

// IsSectionBoundary reports whether the token starts or ends a section.
func (t Token) IsSectionBoundary() bool {
	return t.Type == TokenSectionStart || t.Type == TokenSectionEnd
}

// IsContentToken reports whether the token carries human-readable text
// (as opposed to purely structural/metadata information).
func (t Token) IsContentToken() bool {
	switch t.Type {
	case TokenText, TokenHeading, TokenTitle, TokenExcerpt:
		return true
	default:
		return false
	}
}

// IsMetadata reports whether the token holds metadata key/values.
func (t Token) IsMetadata() bool {
	return t.Type == TokenMeta || t.Type == TokenDocumentInfo
}
