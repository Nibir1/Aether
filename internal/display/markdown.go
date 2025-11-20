// internal/display/markdown.go
//
// This file implements Aether’s Markdown renderer. It converts a
// normalized model.Document into themed Markdown using the Theme
// definitions in themes.go.
//
// Markdown rendering is intentionally pure-text, stable, and deterministic.
// It supports:
//
//   - Title and excerpt
//   - Document metadata
//   - Top-level content paragraphs
//   - Structured sections (article body, feed items, metadata blocks)
//
// This file forms the backbone of Aether's human-facing outputs, including
// CLI previews, TUI views, web displays, and LLM prompt rendering.

package display

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// Renderer converts normalized model.Document instances into Markdown text.
// Each renderer instance is theme-bound.
type Renderer struct {
	theme Theme
}

// NewRenderer creates a new renderer using the given theme.
func NewRenderer(theme Theme) *Renderer {
	return &Renderer{theme: theme}
}

// RenderMarkdown converts a normalized document into Markdown.
// This is the primary entrypoint for Presentation-layer rendering.
func (r *Renderer) RenderMarkdown(doc *model.Document) string {
	if doc == nil {
		return ""
	}

	var out strings.Builder

	// 1. Title
	if strings.TrimSpace(doc.Title) != "" {
		out.WriteString(r.theme.HeadingPrefix)
		out.WriteString(escapeMarkdown(doc.Title))
		out.WriteString("\n\n")
	}

	// 2. Excerpt
	if strings.TrimSpace(doc.Excerpt) != "" {
		out.WriteString("> ")
		out.WriteString(escapeMarkdown(doc.Excerpt))
		out.WriteString("\n\n")
	}

	// 3. Metadata table-like list
	if len(doc.Metadata) > 0 {
		for k, v := range doc.Metadata {
			out.WriteString(r.theme.MetadataPrefix)
			out.WriteString(escapeMarkdown(k))
			out.WriteString(": ")
			out.WriteString(escapeMarkdown(v))
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	// 4. Raw content paragraphs
	if strings.TrimSpace(doc.Content) != "" {
		paras := splitParagraphs(doc.Content)
		for _, p := range paras {
			out.WriteString(escapeMarkdown(p))
			out.WriteString(r.theme.ParagraphSpacing)
		}
		out.WriteString("\n")
	}

	// 5. Structured sections (article body, feed items, etc.)
	for _, sec := range doc.Sections {
		out.WriteString(r.theme.SectionDivider)

		heading := sec.Heading
		if heading == "" {
			// fallback to role when heading missing
			heading = string(sec.Role)
		}

		out.WriteString(r.theme.HeadingPrefix)
		out.WriteString(escapeMarkdown(heading))
		out.WriteString("\n\n")

		// Section text paragraphs
		if strings.TrimSpace(sec.Text) != "" {
			paras := splitParagraphs(sec.Text)
			for _, p := range paras {
				out.WriteString(escapeMarkdown(p))
				out.WriteString(r.theme.ParagraphSpacing)
			}
		}

		// Section metadata
		for k, v := range sec.Meta {
			out.WriteString(r.theme.MetadataPrefix)
			out.WriteString(escapeMarkdown(k))
			out.WriteString(": ")
			out.WriteString(escapeMarkdown(v))
			out.WriteString("\n")
		}

		out.WriteString("\n")
	}

	return strings.TrimSpace(out.String())
}

// splitParagraphs breaks text into paragraphs using double-newline
// boundaries while trimming surrounding whitespace.
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

// escapeMarkdown escapes minimal Markdown characters to avoid structural breakage.
// We deliberately avoid over-escaping to preserve as much original text as possible.
func escapeMarkdown(s string) string {
	replacements := []struct {
		old string
		new string
	}{
		{"#", "\\#"},
		{"*", "\\*"},
		{"_", "\\_"},
	}

	for _, r := range replacements {
		s = strings.ReplaceAll(s, r.old, r.new)
	}

	return s
}
