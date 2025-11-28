// aether/adapter.go
//
// Unified document adapter helpers for Aether.
//
// This file centralizes conversions between:
//   • internal/model.Document        (canonical normalized document)
//   • plugins.Document               (public plugin-facing model)
//
// Improvements in this version:
//   • Added safe namespaced metadata keys: aether.url, aether.source
//   • Added SourceURL override from plugin.Metadata["aether.url"] if present
//   • Applied whitespace normalization (TrimSpace) consistently
//   • Preserved plugin.Source safely without colliding with user metadata
//   • More defensive copying + cleaner logic
//
// These helpers are used throughout Aether’s TransformPlugins pipeline,
// DisplayPlugins, and any adapter-style integrations.

package aether

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/plugins"
)

//
// ───────────────────────────────────────────────────────────────────────────
//                     MODEL → PLUGIN DOCUMENT (internal → public)
// ───────────────────────────────────────────────────────────────────────────
//

// modelToPluginDocument converts a normalized model.Document into a
// plugins.Document for TransformPlugins / DisplayPlugins.
func modelToPluginDocument(doc *model.Document) *plugins.Document {
	if doc == nil {
		return &plugins.Document{}
	}

	// --- Clone base metadata ---
	meta := cloneStringMap(doc.Metadata)
	if meta == nil {
		meta = make(map[string]string)
	}

	// --- Ensure canonical URL is represented ---
	if doc.SourceURL != "" {
		if _, exists := meta["aether.url"]; !exists {
			meta["aether.url"] = doc.SourceURL
		}
	}

	// --- Stable "source" tag if none provided ---
	source := meta["aether.source"]
	if source == "" {
		source = "aether:normalized"
	}

	// --- Normalize text fields ---
	title := strings.TrimSpace(doc.Title)
	excerpt := strings.TrimSpace(doc.Excerpt)
	content := strings.TrimSpace(doc.Content)

	p := &plugins.Document{
		Source:   source,
		URL:      doc.SourceURL,
		Kind:     plugins.DocumentKind(doc.Kind),
		Title:    title,
		Excerpt:  excerpt,
		Content:  content,
		Metadata: meta,
		Sections: make([]plugins.Section, 0, len(doc.Sections)),
	}

	// --- Convert sections ---
	for _, s := range doc.Sections {
		p.Sections = append(p.Sections, plugins.Section{
			Role:  plugins.SectionRole(s.Role),
			Title: strings.TrimSpace(s.Heading),
			Text:  strings.TrimSpace(s.Text),
			Meta:  cloneStringMap(s.Meta),
		})
	}

	return p
}

//
// ───────────────────────────────────────────────────────────────────────────
//                     PLUGIN → MODEL DOCUMENT (public → internal)
// ───────────────────────────────────────────────────────────────────────────
//

// pluginToModelDocument converts a plugin-generated Document back into
// the canonical internal model.Document.
func pluginToModelDocument(pdoc *plugins.Document) *model.Document {
	if pdoc == nil {
		return nil
	}

	// --- Clone metadata ---
	meta := cloneStringMap(pdoc.Metadata)
	if meta == nil {
		meta = make(map[string]string)
	}

	// --- Inject plugin source reliably under namespaced key ---
	if pdoc.Source != "" {
		if _, exists := meta["aether.source"]; !exists {
			meta["aether.source"] = pdoc.Source
		}
	}

	// --- Assign canonical SourceURL ---
	sourceURL := strings.TrimSpace(pdoc.URL)

	// Override if metadata explicitly includes a canonical URL
	if u, ok := meta["aether.url"]; ok && strings.TrimSpace(u) != "" {
		sourceURL = strings.TrimSpace(u)
	}

	// --- Normalize text fields ---
	title := strings.TrimSpace(pdoc.Title)
	excerpt := strings.TrimSpace(pdoc.Excerpt)
	content := strings.TrimSpace(pdoc.Content)

	out := &model.Document{
		SourceURL: sourceURL,
		Kind:      model.DocumentKind(pdoc.Kind),
		Title:     title,
		Excerpt:   excerpt,
		Content:   content,
		Metadata:  meta,
		Sections:  make([]model.Section, 0, len(pdoc.Sections)),
	}

	// --- Convert sections ---
	for _, s := range pdoc.Sections {
		out.Sections = append(out.Sections, model.Section{
			Role:    model.SectionRole(s.Role),
			Heading: strings.TrimSpace(s.Title),
			Text:    strings.TrimSpace(s.Text),
			Meta:    cloneStringMap(s.Meta),
		})
	}

	return out
}

//
// ───────────────────────────────────────────────────────────────────────────
//                             MAP HELPERS
// ───────────────────────────────────────────────────────────────────────────
//

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
