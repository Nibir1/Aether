// aether/display.go
//
// Public wrappers around Aether’s display subsystem.
//
// This file exposes a stable API for:
//   • Markdown rendering
//   • Preview rendering
//   • Table rendering
//   • Theme selection
//   • DisplayPlugin routing (Stage 2)
//
// Internally all heavy logic lives inside internal/display.

package aether

import (
	"context"
	"fmt"

	"github.com/Nibir1/Aether/internal/display"
	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/plugins"
)

//
// ───────────────────────────────────────────────────────────────────────────
//                              MARKDOWN RENDERING
// ───────────────────────────────────────────────────────────────────────────
//

// RenderMarkdown renders a normalized document with the default theme.
func (c *Client) RenderMarkdown(doc *NormalizedDocument) string {
	r := display.NewRenderer(display.DefaultTheme())
	return r.RenderMarkdown((*model.Document)(doc))
}

// RenderMarkdownWithTheme renders a normalized document with a custom theme.
func (c *Client) RenderMarkdownWithTheme(doc *NormalizedDocument, theme display.Theme) string {
	r := display.NewRenderer(theme)
	return r.RenderMarkdown((*model.Document)(doc))
}

//
// ───────────────────────────────────────────────────────────────────────────
//                               PREVIEW RENDERING
// ───────────────────────────────────────────────────────────────────────────
//

// Previews are short summaries containing:
//   • Title
//   • Excerpt / summary
//   • First paragraph fallback

func (c *Client) RenderPreview(doc *NormalizedDocument) string {
	pr := display.NewPreviewRenderer(display.DefaultTheme())
	p := pr.MakePreview((*model.Document)(doc))
	return pr.RenderPreview(p)
}

func (c *Client) RenderPreviewWithTheme(doc *NormalizedDocument, theme display.Theme) string {
	pr := display.NewPreviewRenderer(theme)
	p := pr.MakePreview((*model.Document)(doc))
	return pr.RenderPreview(p)
}

//
// ───────────────────────────────────────────────────────────────────────────
//                                 TABLE RENDERING
// ───────────────────────────────────────────────────────────────────────────
//

func (c *Client) RenderTable(header []string, rows [][]string) string {
	tbl := display.Table{Header: header, Rows: rows}
	return display.RenderTable(display.DefaultTheme(), tbl)
}

func (c *Client) RenderTableWithTheme(header []string, rows [][]string, theme display.Theme) string {
	tbl := display.Table{Header: header, Rows: rows}
	return display.RenderTable(theme, tbl)
}

//
// ───────────────────────────────────────────────────────────────────────────
//                                  PUBLIC THEMES
// ───────────────────────────────────────────────────────────────────────────
//

func DefaultTheme() display.Theme { return display.DefaultTheme() }
func DarkTheme() display.Theme    { return display.DarkTheme() }
func MinimalTheme() display.Theme { return display.MinimalTheme() }
func PaperTheme() display.Theme   { return display.PaperTheme() }

//
// ───────────────────────────────────────────────────────────────────────────
//                                 NORMALIZATION
// ───────────────────────────────────────────────────────────────────────────
//

// ToNormalized converts a SearchResult to a normalized model.Document.
func (c *Client) ToNormalized(sr *SearchResult) *model.Document {
	return c.NormalizeSearchResult(sr)
}

//
// ───────────────────────────────────────────────────────────────────────────
//                     DISPLAY PLUGIN ROUTING (NEW — STEP 2)
// ───────────────────────────────────────────────────────────────────────────
//

// GetDisplayPlugin returns a display plugin that supports a given format tag.
// Example: GetDisplayPlugin("html"), GetDisplayPlugin("pdf").
func (c *Client) GetDisplayPlugin(format string) (plugins.DisplayPlugin, bool) {
	if c == nil || c.plugins == nil {
		return nil, false
	}
	p := c.plugins.FindDisplayByFormat(format)
	if p == nil {
		return nil, false
	}
	return p, true
}

// ListDisplayFormats returns all registered display formats.
// Example output: []string{"html", "ansi", "pdf"}
func (c *Client) ListDisplayFormats() []string {
	if c == nil || c.plugins == nil {
		return nil
	}
	return c.plugins.ListDisplayFormats()
}

// RenderWithPlugin renders a normalized document using a DisplayPlugin.
// The plugin determines the output format (text, html, pdf, ansi, etc.).
func (c *Client) RenderWithPlugin(ctx context.Context, format string, doc *NormalizedDocument) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("aether: nil client")
	}
	if doc == nil {
		return nil, fmt.Errorf("aether: nil document")
	}

	p, ok := c.GetDisplayPlugin(format)
	if !ok {
		return nil, fmt.Errorf("aether: no display plugin for format %q", format)
	}

	// Convert normalized → plugin.Document
	pdoc := c.toPluginDocument(doc)
	return p.Render(ctx, pdoc)
}

// RenderSearchResultWithPlugin normalizes a SearchResult and renders it using a plugin.
func (c *Client) RenderSearchResultWithPlugin(ctx context.Context, format string, sr *SearchResult) ([]byte, error) {
	if sr == nil {
		return nil, fmt.Errorf("aether: nil SearchResult")
	}
	doc := c.NormalizeSearchResult(sr)
	return c.RenderWithPlugin(ctx, format, doc)
}

//
// ───────────────────────────────────────────────────────────────────────────
//             INTERNAL ADAPTER: model.Document → plugins.Document
// ───────────────────────────────────────────────────────────────────────────
//

// toPluginDocument converts a normalized model.Document into a plugins.Document
// so DisplayPlugins can render a stable public structure.
func (c *Client) toPluginDocument(doc *model.Document) *plugins.Document {
	if doc == nil {
		return &plugins.Document{}
	}

	// Convert top-level fields
	p := &plugins.Document{
		Source:   "aether-normalized",
		URL:      doc.SourceURL,
		Title:    doc.Title,
		Excerpt:  doc.Excerpt,
		Content:  doc.Content,
		Kind:     plugins.DocumentKind(doc.Kind),
		Metadata: map[string]string{},
		Sections: make([]plugins.Section, 0, len(doc.Sections)),
	}

	// Copy metadata
	for k, v := range doc.Metadata {
		p.Metadata[k] = v
	}

	// Convert sections
	for _, s := range doc.Sections {
		p.Sections = append(p.Sections, plugins.Section{
			Role:  plugins.SectionRole(s.Role),
			Title: s.Heading,
			Text:  s.Text,
			Meta:  cloneMeta(s.Meta),
		})
	}

	return p
}

func cloneMeta(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
