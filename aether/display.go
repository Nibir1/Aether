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
// The Model ↔ Plugin adapter logic lives in adapter.go.

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

func normalizeFormat(f string) string {
	return strings.ToLower(strings.TrimSpace(f))
}

func (c *Client) FindDisplayPlugin(format string) (plugins.DisplayPlugin, bool) {
	if c == nil || c.plugins == nil {
		return nil, false
	}
	p := c.plugins.FindDisplayByFormat(normalizeFormat(format))
	if p == nil {
		return nil, false
	}
	return p, true
}

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

// Render renders a normalized document into a given format.
// Built-in: markdown/md, preview, text
// All other formats → MUST come from a DisplayPlugin.
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

	// ───── Plugin-required formats (Strict Mode) ───────────────────────────
	if c.plugins == nil {
		return nil, fmt.Errorf("aether: no plugin registry available for format %q", f)
	}

	p := c.plugins.FindDisplayByFormat(f)
	if p == nil {
		return nil, fmt.Errorf("aether: no display plugin registered for format %q", f)
	}

	// NormalizedDocument is an alias of model.Document — convert properly.
	m := (*model.Document)(doc)
	pdoc := modelToPluginDocument(m)

	return p.Render(ctx, pdoc)
}

// RenderSearchResult normalizes a SearchResult and passes it to Render().
func (c *Client) RenderSearchResult(ctx context.Context, format string, sr *SearchResult) ([]byte, error) {
	if sr == nil {
		return nil, fmt.Errorf("aether: nil SearchResult")
	}
	doc := c.NormalizeSearchResult(sr)
	return c.Render(ctx, format, doc)
}
