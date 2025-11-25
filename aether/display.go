// aether/display.go
//
// Public wrappers around Aether’s display subsystem.
//
// This file exposes a stable API for:
//   • Markdown rendering
//   • Preview rendering
//   • Table rendering
//   • Theme selection
//
// Internally all heavy logic lives inside internal/display.

package aether

import (
	"github.com/Nibir1/Aether/internal/display"
	"github.com/Nibir1/Aether/internal/model"
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
//
// PreviewRenderer is separate from Renderer.

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
// Tables are rendered using the standalone RenderTable(...) function.

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
