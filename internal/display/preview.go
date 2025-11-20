// internal/display/preview.go
//
// This file defines Aether’s preview utilities. Previews are short,
// human-readable summaries or snippets derived from normalized documents.
//
// They are used by:
//   - Search result lists
//   - CLI interfaces (list view)
//   - TUI view components
//   - LLM prompt preparation
//
// Previews follow three goals:
//
//   1. Brevity       — only the most important parts of a document.
//   2. Determinism   — same input produces same preview.
//   3. Safety        — no expensive parsing; relies on pre-normalized data.
//
// This module does NOT render Markdown; it produces plain text snippets.
// (Markdown rendering happens in markdown.go.)

package display

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// Preview contains a compact representation of a document for listing
// or summary purposes.
type Preview struct {
	Title    string
	Excerpt  string
	Snippet  string // 1–3 meaningful body paragraphs
	Metadata map[string]string
}

// BuildPreview creates a short preview suitable for CLI/TUI search results,
// list pages, or summary displays.
//
// Behavior:
//   - Uses Title → Excerpt → first paragraphs → first section paragraphs
//   - Strips whitespace
//   - Truncates long text intelligently
//   - Ensures deterministic output
//
// This function never panics and is safe to call with nil.
func BuildPreview(doc *model.Document) Preview {
	if doc == nil {
		return Preview{
			Title:    "",
			Excerpt:  "",
			Snippet:  "",
			Metadata: map[string]string{},
		}
	}

	rm := BuildRenderModel(doc) // convert to display model

	preview := Preview{
		Title:    rm.Title,
		Excerpt:  rm.Excerpt,
		Metadata: map[string]string{},
	}

	// Copy metadata only for keys that are useful for previewing.
	for k, v := range rm.Metadata {
		preview.Metadata[k] = v
	}

	// Build snippet: try body paragraphs first.
	snippet := ""

	if len(rm.BodyParagraphs) > 0 {
		snippet = rm.BodyParagraphs[0]

		if len(rm.BodyParagraphs) > 1 {
			snippet = snippet + " " + rm.BodyParagraphs[1]
		}
	} else {
		// fallback to first section paragraphs
		for _, sec := range rm.Sections {
			if len(sec.Paragraphs) > 0 {
				snippet = sec.Paragraphs[0]
				if len(sec.Paragraphs) > 1 {
					snippet = snippet + " " + sec.Paragraphs[1]
				}
				break
			}
		}
	}

	preview.Snippet = cleanAndTruncate(snippet, 320) // soft truncation

	return preview
}

// cleanAndTruncate collapses whitespace and truncates to maxLen characters,
// adding a “…” suffix when truncation occurs.
func cleanAndTruncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = collapseWhitespace(s)

	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

// collapseWhitespace replaces runs of whitespace with a single space.
func collapseWhitespace(s string) string {
	out := make([]rune, 0, len(s))
	space := false

	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			if !space {
				out = append(out, ' ')
			}
			space = true
		} else {
			out = append(out, r)
			space = false
		}
	}

	return strings.TrimSpace(string(out))
}
