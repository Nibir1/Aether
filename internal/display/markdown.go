// internal/display/markdown.go
//
// Markdown transformation utilities for Aether’s Display subsystem.
//
// This file provides the foundational building blocks used by
// model_render.go to convert normalized documents into theme-aware
// Markdown or plain text.
//
// Features:
//   • Heading rendering with theme rules + optional ANSI color
//   • Bullet lists with theme-configurable prefix
//   • Optional line wrapping according to theme.MaxWidth / EffectiveWidth
//   • Paragraph normalization and trimming
//   • Inline emphasis helpers
//
// This renderer avoids “HTML in Markdown” and stays within the
// CommonMark-safe subset so output is compatible with GitHub,
// terminals, chatbots, and LLMs.

package display

import (
	"strings"
	"unicode"
)

// RenderHeading renders a Markdown heading (level 1–6) using the theme.
// It applies:
//   - optional ANSI styling (color.go)
//   - configurable heading style (prefix, underline, uppercase)
//   - safe trimming of whitespace
func RenderHeading(t Theme, level int, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	style := t.HeadingForLevel(level)

	headingText := text
	if style.Uppercase {
		headingText = strings.ToUpper(headingText)
	}

	var out strings.Builder

	// Prefix-based style (e.g. "# ", "## ", "• ")
	if style.Prefix != "" {
		out.WriteString(style.Prefix)
		out.WriteString(headingText)
	} else {
		out.WriteString(headingText)
	}

	rendered := out.String()

	// Apply ANSI styling if theme allows color.
	rendered = styleHeading(t, rendered)

	// Optional underline style:
	//   Title
	//   =====
	if style.Underline {
		underlineRune := style.UnderlineRune
		if underlineRune == 0 {
			underlineRune = '='
		}
		underline := strings.Repeat(string(underlineRune), len(headingText))
		rendered = rendered + "\n" + underline
	}

	return rendered
}

// RenderParagraph formats a block of prose text. This includes:
//
//   - Normalizing whitespace
//   - Optional wrapping to the effective width (theme + terminal)
//   - Optional ANSI emphasis (e.g., strong/em) added later in pipeline
//
// The function preserves paragraph structure but removes excessive
// internal spacing.
func RenderParagraph(t Theme, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Normalize internal whitespace: convert multiple spaces to single.
	text = collapseSpaces(text)

	// Wrap to effective width (theme or terminal).
	width := EffectiveWidth(t)
	return wrapText(text, width)
}

// RenderBulletList formats a slice of bullet items using the theme’s
// Bullet marker. Each item is rendered as a wrapped paragraph prefixed
// by the bullet symbol.
func RenderBulletList(t Theme, items []string) string {
	var b strings.Builder
	bullet := t.Bullet
	if bullet == "" {
		bullet = "-"
	}

	width := EffectiveWidth(t)

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		line := bullet + " " + item
		line = wrapText(line, width)

		b.WriteString(line)
		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}

// RenderCodeBlock wraps code in a Markdown fenced block.
//
// Example:
//
// ```
// line1
// line2
// ```
func RenderCodeBlock(code string) string {
	code = strings.TrimRight(code, "\n")
	return "```\n" + code + "\n```"
}

// collapseSpaces normalizes internal whitespace sequences.
// It preserves intentional newlines but avoids text looking broken.
func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	spaceSeen := false
	for _, r := range s {
		if unicode.IsSpace(r) && r != '\n' {
			if !spaceSeen {
				b.WriteByte(' ')
				spaceSeen = true
			}
		} else {
			b.WriteRune(r)
			spaceSeen = false
		}
	}
	return b.String()
}

//
//────────────────────────────────────────────
//           TEXT WRAPPING UTILITIES
//────────────────────────────────────────────
//

// wrapText performs simple greedy line wrapping to a given width.
// It preserves words and does not break long tokens.
//
// This avoids external dependencies and is adequate for Aether output.
func wrapText(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var out strings.Builder
	current := ""

	for _, w := range words {
		if len(current)+len(w)+1 > width {
			out.WriteString(strings.TrimSpace(current))
			out.WriteByte('\n')
			current = w
		} else {
			if current == "" {
				current = w
			} else {
				current += " " + w
			}
		}
	}

	if current != "" {
		out.WriteString(strings.TrimSpace(current))
	}

	return out.String()
}

// RenderInlineStrong applies Markdown strong formatting **...**
// and then uses ANSI styling when enabled by the theme.
func RenderInlineStrong(t Theme, s string) string {
	if strings.TrimSpace(s) == "" {
		return s
	}
	return styleStrong(t, "**"+s+"**")
}

// RenderInlineEm applies Markdown emphasis formatting *...*
// and then uses ANSI styling when enabled by the theme.
func RenderInlineEm(t Theme, s string) string {
	if strings.TrimSpace(s) == "" {
		return s
	}
	return styleEm(t, "*"+s+"*")
}

// RenderInlineCode renders inline code using Markdown backticks
// and applies subtle ANSI styling when enabled.
func RenderInlineCode(t Theme, s string) string {
	if strings.TrimSpace(s) == "" {
		return s
	}
	return styleCode(t, "`"+s+"`")
}
