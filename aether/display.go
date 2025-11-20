// aether/display.go
//
// Public wrapper around the internal Markdown renderer.

package aether

import (
	"github.com/Nibir1/Aether/internal/display"
)

// RenderMarkdown renders a normalized document using the default theme.
func (c *Client) RenderMarkdown(doc *NormalizedDocument) string {
	r := display.NewRenderer(display.DefaultTheme)
	return r.RenderMarkdown(doc)
}

// RenderMarkdownWithTheme renders a normalized document with a specific theme.
func (c *Client) RenderMarkdownWithTheme(doc *NormalizedDocument, theme display.Theme) string {
	r := display.NewRenderer(theme)
	return r.RenderMarkdown(doc)
}
