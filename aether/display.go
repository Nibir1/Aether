// aether/display.go
//
// Public wrappers around Aether’s display subsystem.
//
// This file exposes a stable API for:
//   • Markdown rendering (built-in)
//   • Preview rendering (built-in)
//   • Table rendering (built-in)
//   • Theme selection
//   • DisplayPlugin routing (strict mode)
//
// Internally all heavy logic lives inside internal/display.

package aether

import (
	"context"
	"fmt"
	"strings"

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

// Preview is derived from:
//   - Title
//   - Excerpt / summary
//   - First paragraph (fallback)
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
//                     DISPLAY PLUGIN ROUTING (STRICT — OPTION B)
// ───────────────────────────────────────────────────────────────────────────
//

// normalizeFormat ensures format matching is case-insensitive.
func normalizeFormat(f string) string {
	return strings.ToLower(strings.TrimSpace(f))
}

// FindDisplayPlugin returns a plugin by format.
func (c *Client) FindDisplayPlugin(format string) (plugins.DisplayPlugin, bool) {
	if c == nil || c.plugins == nil {
		return nil, false
	}
	p := c.plugins.FindDisplayByFormat(normalizeFormat(format))
	return p, p != nil
}

// ListDisplayFormats lists all registered plugin-provided formats.
func (c *Client) ListDisplayFormats() []string {
	if c == nil || c.plugins == nil {
		return nil
	}
	return c.plugins.ListDisplayFormats()
}

//
// ───────────────────────────────────────────────────────────────────────────
//                   UNIFIED RENDER DISPATCHER (BUILT-IN + PLUGIN)
// ───────────────────────────────────────────────────────────────────────────
//

// Render renders a normalized document using either built-in formats
// or a DisplayPlugin. (Strict Mode)
//
// Built-in formats:
//   - "markdown", "md"
//   - "text" (alias of markdown)
//   - "preview"
//
// All other formats MUST come from DisplayPlugins.
// If no plugin is found → error.
func (c *Client) Render(ctx context.Context, format string, doc *NormalizedDocument) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("aether: nil client")
	}
	if doc == nil {
		return nil, fmt.Errorf("aether: nil document")
	}

	f := normalizeFormat(format)

	// ───── Built-in formats ────────────────────────────────────────────────
	switch f {
	case "markdown", "md", "":
		return []byte(c.RenderMarkdown(doc)), nil

	case "text":
		return []byte(c.RenderMarkdown(doc)), nil

	case "preview":
		return []byte(c.RenderPreview(doc)), nil
	}

	// ───── Plugin formats ──────────────────────────────────────────────────
	p := c.plugins.FindDisplayByFormat(f)
	if p == nil {
		return nil, fmt.Errorf("aether: no display plugin registered for format %q", f)
	}

	pdoc := c.toPluginDocument(doc)
	return p.Render(ctx, pdoc)
}

// RenderSearchResult normalizes a SearchResult then renders it.
func (c *Client) RenderSearchResult(ctx context.Context, format string, sr *SearchResult) ([]byte, error) {
	if sr == nil {
		return nil, fmt.Errorf("aether: nil SearchResult")
	}
	doc := c.NormalizeSearchResult(sr)
	return c.Render(ctx, format, doc)
}

//
// ───────────────────────────────────────────────────────────────────────────
//             INTERNAL ADAPTER: model.Document → plugins.Document
// ───────────────────────────────────────────────────────────────────────────
//

func (c *Client) toPluginDocument(doc *model.Document) *plugins.Document {
	if doc == nil {
		return &plugins.Document{}
	}

	// The unified adapter is aligned with aether/normalize.go
	p := &plugins.Document{
		Source:   "aether-normalized",
		URL:      doc.SourceURL, // canonical source
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

	// Normalize canonical URL into metadata
	if doc.SourceURL != "" {
		p.Metadata["source_url"] = doc.SourceURL
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

// cloneMeta duplicates metadata maps.
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
