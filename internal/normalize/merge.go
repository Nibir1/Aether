// internal/normalize/merge.go
//
// Merging logic for Aether's normalization pipeline.
//
// Each schema_* normalizer (search, article, feed, entity) produces a
// partial model.Document. This file defines mergeDocuments(), which
// combines those partials into a single canonical Document.
//
// Merge strategy:
//
//   • The first non-nil document becomes the base.
//   • Subsequent documents enrich the base without discarding information.
//   • Title/Excerpt/Content are only filled if the base fields are empty.
//   • DocumentKind is preserved unless it is Unknown, in which case a more
//     specific kind from later documents can override it.
//   • Metadata is merged; existing keys on the base are not overwritten.
//   • Sections from all documents are appended in order.
//   • SourceURL is taken from the first document that provides it.

package normalize

import (
	"github.com/Nibir1/Aether/internal/model"
)

// mergeDocuments merges zero or more partial Documents into a single
// canonical Document.
//
// If no non-nil documents are provided, an empty Document is returned.
func mergeDocuments(docs ...*model.Document) *model.Document {
	var base *model.Document

	for _, d := range docs {
		if d == nil {
			continue
		}

		if base == nil {
			// Deep copy first non-nil document as base to avoid
			// mutating the original partial.
			base = copyDocument(d)
			continue
		}

		mergeInto(base, d)
	}

	if base == nil {
		return emptyDocument()
	}
	return base
}

// mergeInto enriches base with data from overlay.
//
// It modifies base in place.
func mergeInto(base, overlay *model.Document) {
	if base == nil || overlay == nil {
		return
	}

	// Kind: upgrade Unknown to something more specific if available.
	if base.Kind == model.DocumentKindUnknown && overlay.Kind != model.DocumentKindUnknown {
		base.Kind = overlay.Kind
	}

	// SourceURL: use first non-empty SourceURL.
	if base.SourceURL == "" && overlay.SourceURL != "" {
		base.SourceURL = overlay.SourceURL
	}

	// Title/Excerpt/Content fallbacks: only fill if base is missing.
	if base.Title == "" && overlay.Title != "" {
		base.Title = overlay.Title
	}
	if base.Excerpt == "" && overlay.Excerpt != "" {
		base.Excerpt = overlay.Excerpt
	}
	if base.Content == "" && overlay.Content != "" {
		base.Content = overlay.Content
	}

	// Metadata: copy missing keys from overlay into base.
	if overlay.Metadata != nil {
		if base.Metadata == nil {
			base.Metadata = make(map[string]string, len(overlay.Metadata))
		}
		for k, v := range overlay.Metadata {
			if _, exists := base.Metadata[k]; !exists {
				base.Metadata[k] = v
			}
		}
	}

	// Sections: append overlay sections (deep copy).
	if len(overlay.Sections) > 0 {
		base.Sections = append(base.Sections, copySections(overlay.Sections)...)
	}
}

// copyDocument performs a deep copy of a Document.
func copyDocument(src *model.Document) *model.Document {
	if src == nil {
		return nil
	}

	out := &model.Document{
		SourceURL: src.SourceURL,
		Kind:      src.Kind,
		Title:     src.Title,
		Excerpt:   src.Excerpt,
		Content:   src.Content,
	}

	// Metadata
	if src.Metadata != nil {
		out.Metadata = copyMetadata(src.Metadata)
	}

	// Sections
	if len(src.Sections) > 0 {
		out.Sections = copySections(src.Sections)
	}

	return out
}

// copySections deep-copies a slice of Sections, including their metadata.
func copySections(src []model.Section) []model.Section {
	if len(src) == 0 {
		return nil
	}

	out := make([]model.Section, 0, len(src))
	for _, s := range src {
		ns := model.Section{
			Role:    s.Role,
			Heading: s.Heading,
			Text:    s.Text,
		}
		if s.Meta != nil {
			ns.Meta = copyMetadata(s.Meta)
		}
		out = append(out, ns)
	}
	return out
}
