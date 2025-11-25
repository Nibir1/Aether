// internal/toon/builder.go
//
// Token builder DSL for TOON documents.
//
// This helper struct makes it easier to construct well-formed token
// streams from higher-level structures (e.g., model.Document) while
// keeping the grammar rules in a single place.

package toon

// Builder accumulates TOON tokens in order.
type Builder struct {
	tokens []Token
}

// NewBuilder creates an empty token builder.
func NewBuilder() *Builder {
	return &Builder{
		tokens: make([]Token, 0, 32),
	}
}

// Tokens returns a copy of the accumulated tokens.
func (b *Builder) Tokens() []Token {
	out := make([]Token, len(b.tokens))
	copy(out, b.tokens)
	return out
}

// DocumentInfo emits a DOCINFO token with basic attributes such as
// document kind and optional extra attributes.
func (b *Builder) DocumentInfo(kind string, attrs map[string]string) {
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["kind"] = kind

	b.tokens = append(b.tokens, Token{
		Type:  TokenDocumentInfo,
		Attrs: attrs,
	})
}

// Title emits a TITLE-like heading token (optional).
func (b *Builder) Title(text string) {
	if text == "" {
		return
	}
	b.tokens = append(b.tokens, Token{
		Type: TokenTitle,
		Role: "title",
		Text: text,
	})
}

// Excerpt emits an EXCERPT token.
func (b *Builder) Excerpt(text string) {
	if text == "" {
		return
	}
	b.tokens = append(b.tokens, Token{
		Type: TokenExcerpt,
		Role: "excerpt",
		Text: text,
	})
}

// SectionStart emits a SECTION_START token with role and optional
// heading.
func (b *Builder) SectionStart(role, heading string) {
	attrs := map[string]string{}
	if heading != "" {
		attrs["heading"] = heading
	}
	if role != "" {
		attrs["role"] = role
	}

	b.tokens = append(b.tokens, Token{
		Type:  TokenSectionStart,
		Role:  role,
		Attrs: attrs,
	})
}

// SectionEnd emits a SECTION_END token for the given role.
func (b *Builder) SectionEnd(role string) {
	b.tokens = append(b.tokens, Token{
		Type: TokenSectionEnd,
		Role: role,
	})
}

// Heading emits a HEADING token inside a section.
func (b *Builder) Heading(role, text string) {
	if text == "" {
		return
	}
	b.tokens = append(b.tokens, Token{
		Type: TokenHeading,
		Role: role,
		Text: text,
	})
}

// TextBlock emits a TEXT token inside a section or as standalone
// content.
func (b *Builder) TextBlock(role, text string) {
	if text == "" {
		return
	}
	b.tokens = append(b.tokens, Token{
		Type: TokenText,
		Role: role,
		Text: text,
	})
}

// MetaKV emits a META token with a single key/value pair.
func (b *Builder) MetaKV(role, key, value string) {
	if key == "" || value == "" {
		return
	}
	b.tokens = append(b.tokens, Token{
		Type: TokenMeta,
		Role: role,
		Attrs: map[string]string{
			key: value,
		},
	})
}
