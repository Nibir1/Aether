// internal/display/preview.go
//
// Preview rendering for Aether’s Display subsystem.
//
// This file provides compact, human-friendly previews of normalized
// model.Document values. Previews are ideal for:
//
//   • CLI search results
//   • TUI lists
//   • Quick summaries
//   • Logging
//
// A preview contains:
//   • Title (if any)
//   • Excerpt (if any)
//   • Or the first non-empty paragraph of body text
//
// Previews respect:
//   • Theme width (EffectiveWidth)
//   • UTF-8 safe truncation
//   • ANSI color rules via styleHeading / styleEm
//
// This renderer does not depend on the full model_render pipeline;
// it is intentionally lightweight for performance and clarity.

package display

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// Preview represents a human-friendly short display representation
// of a normalized document.
type Preview struct {
	Title   string
	Summary string
}

// PreviewRenderer renders Preview structs using a Theme.
type PreviewRenderer struct {
	Theme Theme
}

// NewPreviewRenderer constructs a preview renderer using the given Theme.
func NewPreviewRenderer(t Theme) PreviewRenderer {
	return PreviewRenderer{Theme: sanitizeTheme(t)}
}

// MakePreview extracts a Preview struct from a normalized Document.
//
// Rules:
//  1. Title = doc.Title OR fallback to SourceURL.
//  2. Summary = doc.Excerpt OR first non-empty paragraph from sections or content.
//
// This struct-level function does not perform formatting; renderers do.
func (PreviewRenderer) MakePreview(doc *model.Document) Preview {
	if doc == nil {
		return Preview{}
	}

	// Title selection.
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		title = strings.TrimSpace(doc.SourceURL)
	}

	// Summary selection:
	//   Priority: Excerpt → first section paragraph → Content
	summary := ""
	if strings.TrimSpace(doc.Excerpt) != "" {
		summary = doc.Excerpt
	} else {
		summary = firstNonEmptyParagraph(doc)
	}

	return Preview{
		Title:   title,
		Summary: summary,
	}
}

// RenderPreview produces a human-readable, theme-aware single-block preview
// suitable for CLI or UI list displays.
func (r PreviewRenderer) RenderPreview(p Preview) string {
	if p.Title == "" && p.Summary == "" {
		return ""
	}

	var b strings.Builder

	// Render title
	if p.Title != "" {
		h := styleHeading(r.Theme, p.Title)
		h = wrapTextToWidth(h, EffectiveWidth(r.Theme))
		b.WriteString(h)
	}

	// Render summary
	if p.Summary != "" {
		if p.Title != "" {
			b.WriteByte('\n')
		}
		sum := strings.TrimSpace(p.Summary)
		sum = styleEm(r.Theme, sum)
		sum = wrapTextToWidth(sum, EffectiveWidth(r.Theme))
		b.WriteString(sum)
	}

	return b.String()
}

//
//────────────────────────────────────────────────────────────────────────────
//                              HELPERS
//────────────────────────────────────────────────────────────────────────────
//

// firstNonEmptyParagraph tries to extract the first meaningful paragraph
// from Document.Content or Document.Sections.
func firstNonEmptyParagraph(doc *model.Document) string {
	// 1. Try content paragraphs
	if strings.TrimSpace(doc.Content) != "" {
		par := firstParagraph(doc.Content)
		if par != "" {
			return par
		}
	}

	// 2. Try section bodies
	for _, s := range doc.Sections {
		if strings.TrimSpace(s.Text) != "" {
			par := firstParagraph(s.Text)
			if par != "" {
				return par
			}
		}
	}

	return ""
}

// firstParagraph extracts the first non-empty paragraph (double-newline separated).
func firstParagraph(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	blocks := strings.Split(text, "\n\n")
	for _, b := range blocks {
		b = strings.TrimSpace(b)
		if b != "" {
			return b
		}
	}
	return ""
}
