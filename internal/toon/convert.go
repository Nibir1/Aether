// internal/toon/convert.go
//
// Conversion from model.Document → TOON Document.
// This is the primary entry point used by Aether when marshaling TOON.

package toon

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// FromModel converts a normalized model.Document into a TOON document.
//
// It encodes:
//   - document kind + metadata into Attributes + DOCINFO token
//   - title and excerpt into dedicated tokens
//   - core content into TEXT tokens
//   - sections into SECTION_START / SECTION_END + HEADING + TEXT + META tokens.
func FromModel(m *model.Document) *Document {
	if m == nil {
		return &Document{
			Kind:       model.DocumentKindUnknown,
			Attributes: map[string]string{},
			Tokens:     nil,
		}
	}

	title := strings.TrimSpace(m.Title)
	excerpt := strings.TrimSpace(m.Excerpt)
	content := strings.TrimSpace(m.Content)

	out := &Document{
		SourceURL:  m.SourceURL,
		Kind:       m.Kind,
		Title:      title,
		Excerpt:    excerpt,
		Attributes: cloneMap(m.Metadata),
		Tokens:     nil,
	}

	b := NewBuilder()

	// DocumentInfo token with kind and any high-level metadata of interest.
	b.DocumentInfo(string(m.Kind), nil)

	// Optional title / excerpt tokens.
	if title != "" {
		b.Title(title)
	}
	if excerpt != "" {
		b.Excerpt(excerpt)
	}

	// Optional content token (only when there are no sections, or as
	// a fallback summary).
	if content != "" && len(m.Sections) == 0 {
		b.TextBlock("content", content)
	}

	// Section-based structure.
	for _, sec := range m.Sections {
		role := string(sec.Role)
		heading := strings.TrimSpace(sec.Heading)
		body := strings.TrimSpace(sec.Text)

		// SECTION_START
		b.SectionStart(role, heading)

		// Optional heading token
		if heading != "" {
			b.Heading(role, heading)
		}

		// Body text
		if body != "" {
			b.TextBlock(role, body)
		}

		// Section metadata → META tokens
		for k, v := range sec.Meta {
			b.MetaKV(role, k, v)
		}

		// SECTION_END
		b.SectionEnd(role)
	}

	out.Tokens = b.Tokens()
	return out
}
