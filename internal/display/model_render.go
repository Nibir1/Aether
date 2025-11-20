// internal/display/model_render.go
//
// This file defines a small, display-oriented abstraction over the
// normalized model.Document type. While model.Document is the canonical
// normalized representation used by Aether internally, the Display layer
// often benefits from a slightly different shape:
//
//   - Content already split into paragraphs
//   - Sections flattened into renderable units
//   - Metadata copied into a stable map
//
// RenderModel and its helpers are designed to be shared by multiple
// output backends (Markdown, plain text, ANSI, HTML) without each
// renderer having to repeat paragraph-splitting and section-normalizing
// logic.

package display

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// RenderSection represents a display-ready section derived from
// model.Section. It contains pre-split paragraphs and a stable
// metadata map.
type RenderSection struct {
	Title      string
	Role       model.SectionRole
	Paragraphs []string
	Meta       map[string]string
}

// RenderModel is a display-oriented view of a normalized document.
//
// It is intentionally focused on what renderers need:
//   - Title and excerpt
//   - Stable metadata
//   - Top-level body paragraphs
//   - A list of structured sections, each with paragraphs and metadata
type RenderModel struct {
	Title          string
	Excerpt        string
	Metadata       map[string]string
	BodyParagraphs []string
	Sections       []RenderSection
}

// BuildRenderModel constructs a RenderModel from a normalized document.
//
// It performs paragraph splitting, shallow metadata copying, and
// section normalization. This keeps higher-level renderers simple and
// avoids repeating core transformation logic for each output format.
//
// The function is safe to call with nil; in that case it returns an
// empty RenderModel.
func BuildRenderModel(doc *model.Document) *RenderModel {
	if doc == nil {
		return &RenderModel{
			Title:          "",
			Excerpt:        "",
			Metadata:       map[string]string{},
			BodyParagraphs: nil,
			Sections:       nil,
		}
	}

	rm := &RenderModel{
		Title:    strings.TrimSpace(doc.Title),
		Excerpt:  strings.TrimSpace(doc.Excerpt),
		Metadata: map[string]string{},
	}

	// Copy metadata so renderers can safely mutate if needed.
	for k, v := range doc.Metadata {
		rm.Metadata[k] = v
	}

	// Split top-level content into paragraphs.
	if strings.TrimSpace(doc.Content) != "" {
		rm.BodyParagraphs = splitParagraphs(doc.Content)
	}

	// Normalize sections: split text into paragraphs and copy meta.
	for _, sec := range doc.Sections {
		rs := RenderSection{
			Title:      strings.TrimSpace(sec.Heading),
			Role:       sec.Role,
			Paragraphs: nil,
			Meta:       map[string]string{},
		}

		// If no explicit heading, fall back to role label.
		if rs.Title == "" {
			rs.Title = string(sec.Role)
		}

		if strings.TrimSpace(sec.Text) != "" {
			rs.Paragraphs = splitParagraphs(sec.Text)
		}

		for k, v := range sec.Meta {
			rs.Meta[k] = v
		}

		rm.Sections = append(rm.Sections, rs)
	}

	return rm
}
