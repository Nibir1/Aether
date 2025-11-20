// internal/normalize/schema_entity.go
//
// Normalizes structured entities (Wikidata, Wikipedia Summary, OpenAPI
// responses, GovPress posts, etc.) into model.Section values with the
// role "entity". These structured blocks may come from multiple sources
// within a SearchResult.
//
// Entity normalization rules:
//   • Each entity becomes one section within a wrapping Document.
//   • Heading = entity.Label (fallback: entity.ID).
//   • Text = entity.Summary (fallback: label).
//   • Metadata = ID, URL, and entity.Metadata fields.
//   • merge.go will integrate these into the primary normalized Document.

package normalize

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// normalizeEntities produces a standalone Document with multiple entity
// sections. Each Entity in sr.Entities becomes a SectionRoleEntity.
func normalizeEntities(sr *SearchResult) *model.Document {
	if sr == nil || len(sr.Entities) == 0 {
		return nil
	}

	sections := make([]model.Section, 0, len(sr.Entities))

	for _, e := range sr.Entities {
		if e == nil {
			continue
		}

		heading := strings.TrimSpace(e.Label)
		summary := strings.TrimSpace(e.Summary)

		// Fallback heading
		if heading == "" {
			if e.ID != "" {
				heading = e.ID
			} else {
				heading = "(entity)"
			}
		}

		// Fallback summary
		if summary == "" {
			summary = heading
		}

		// Construct metadata
		meta := map[string]string{}
		if e.ID != "" {
			meta["id"] = strings.TrimSpace(e.ID)
		}
		if e.URL != "" {
			meta["url"] = strings.TrimSpace(e.URL)
		}
		for k, v := range e.Metadata {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if k != "" && v != "" {
				meta[k] = v
			}
		}

		sections = append(sections, model.Section{
			Role:    model.SectionRoleEntity,
			Heading: heading,
			Text:    summary,
			Meta:    meta,
		})
	}

	doc := &model.Document{
		Kind:     model.DocumentKindEntity,
		Title:    deriveEntityTitle(sr),
		Excerpt:  "",
		Content:  "",
		Metadata: map[string]string{},
		Sections: sections,
	}

	return doc
}

//
// ────────────────────────────────────────────────────────────────────────
//                             HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// deriveEntityTitle determines a title for the entire set of entity sections.
func deriveEntityTitle(sr *SearchResult) string {
	if len(sr.Entities) == 1 && sr.Entities[0] != nil {
		// Single-entity case: use the label directly.
		if strings.TrimSpace(sr.Entities[0].Label) != "" {
			return strings.TrimSpace(sr.Entities[0].Label)
		}
		if strings.TrimSpace(sr.Entities[0].ID) != "" {
			return strings.TrimSpace(sr.Entities[0].ID)
		}
	}

	// Multi-entity or unknown case:
	if sr.PrimaryDocument != nil && sr.PrimaryDocument.Title != "" {
		return strings.TrimSpace(sr.PrimaryDocument.Title)
	}

	return "(entities)"
}
