// aether/adapter.go
//
// Unified document adapter helpers for Aether.
//
// This file centralizes conversions between:
//   • internal/model.Document        (canonical normalized document)
//   • plugins.Document               (public plugin-facing model)
//
// These helpers are used by:
//   • TransformPlugins pipeline (Normalize → Transform → Normalize)
//   • DisplayPlugins routing and rendering
//   • Any future adapter-style integrations inside the aether package.
//
// Design goals:
//   • Single source of truth for field mapping
//   • Stable mapping of URL / Source / Kind / Sections / Metadata
//   • No duplication across normalize.go, display.go, toon helpers, etc.

package aether

import (
	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/plugins"
)

//
// ───────────────────────────────────────────────────────────────────────────
//                       MODEL → PLUGIN DOCUMENT
// ───────────────────────────────────────────────────────────────────────────
//

// modelToPluginDocument converts a canonical normalized model.Document into
// a plugins.Document for use by TransformPlugins and DisplayPlugins.
//
// Mapping rules:
//   - SourceURL → URL (plugin)
//   - Kind      → Kind (string enum passthrough)
//   - Title / Excerpt / Content copied directly
//   - Metadata cloned (shallow copy)
//   - Sections: Heading ↔ Title, Role + Meta preserved
//   - If Metadata["url"] is missing but SourceURL is present, it is added.
//   - If Metadata["source"] is missing, a default "aether:normalized" is used.
func modelToPluginDocument(doc *model.Document) *plugins.Document {
	if doc == nil {
		return &plugins.Document{}
	}

	meta := cloneStringMap(doc.Metadata)
	if meta == nil {
		meta = make(map[string]string)
	}

	// Ensure a stable "url" metadata key if we have a SourceURL.
	if doc.SourceURL != "" {
		if _, exists := meta["url"]; !exists {
			meta["url"] = doc.SourceURL
		}
	}

	// Default logical source if none provided.
	source := meta["source"]
	if source == "" {
		source = "aether:normalized"
	}

	p := &plugins.Document{
		Source:   source,
		URL:      doc.SourceURL,
		Kind:     plugins.DocumentKind(doc.Kind),
		Title:    doc.Title,
		Excerpt:  doc.Excerpt,
		Content:  doc.Content,
		Metadata: meta,
		Sections: make([]plugins.Section, 0, len(doc.Sections)),
	}

	for _, s := range doc.Sections {
		p.Sections = append(p.Sections, plugins.Section{
			Role:  plugins.SectionRole(s.Role),
			Title: s.Heading,
			Text:  s.Text,
			Meta:  cloneStringMap(s.Meta),
		})
	}

	return p
}

//
// ───────────────────────────────────────────────────────────────────────────
//                       PLUGIN → MODEL DOCUMENT
// ───────────────────────────────────────────────────────────────────────────
//

// pluginToModelDocument converts a plugins.Document produced by a
// TransformPlugin back into the canonical internal model.Document.
//
// Mapping rules:
//   - URL   → SourceURL
//   - Kind  → Kind
//   - Title / Excerpt / Content copied directly
//   - Metadata cloned
//   - If plugin.Source is non-empty and Metadata["source"] is absent,
//     it is injected into metadata.
//   - Sections: Title ↔ Heading, Role + Meta preserved.
func pluginToModelDocument(pdoc *plugins.Document) *model.Document {
	if pdoc == nil {
		return nil
	}

	meta := cloneStringMap(pdoc.Metadata)
	if meta == nil {
		meta = make(map[string]string)
	}

	// Preserve plugin-level source in metadata if not already present.
	if pdoc.Source != "" {
		if _, exists := meta["source"]; !exists {
			meta["source"] = pdoc.Source
		}
	}

	out := &model.Document{
		SourceURL: pdoc.URL,
		Kind:      model.DocumentKind(pdoc.Kind),
		Title:     pdoc.Title,
		Excerpt:   pdoc.Excerpt,
		Content:   pdoc.Content,
		Metadata:  meta,
		Sections:  make([]model.Section, 0, len(pdoc.Sections)),
	}

	for _, s := range pdoc.Sections {
		out.Sections = append(out.Sections, model.Section{
			Role:    model.SectionRole(s.Role),
			Heading: s.Title,
			Text:    s.Text,
			Meta:    cloneStringMap(s.Meta),
		})
	}

	return out
}

//
// ───────────────────────────────────────────────────────────────────────────
//                              MAP HELPERS
// ───────────────────────────────────────────────────────────────────────────
//

// cloneStringMap safely shallow-copies a map[string]string.
// A nil input returns nil; callers may wrap with defaulting logic.
func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
