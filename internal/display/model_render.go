// internal/display/model_render.go
//
// Rendering of normalized model.Document values into human-readable,
// theme-aware text. This is the main bridge between Aether's internal
// normalized model and the Display subsystem.
//
// The renderer:
//
//   • Produces Markdown-like plain text safe for terminals, logs,
//     and chat interfaces.
//   • Respects Theme width, indentation, color mode, and heading styles.
//   • Treats sections differently based on their SectionRole
//     (body, feed_item, entity, metadata, etc.).
//   • Optionally shows section roles like [feed_item] when
//     Theme.ShowSectionRoles is true.

package display

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// Renderer renders normalized documents according to a Theme.
type Renderer struct {
	Theme Theme
}

// NewRenderer constructs a Renderer using the provided Theme.
//
// The Theme is sanitized to ensure it has sensible defaults (indent,
// bullet, code fence, table padding, etc.).
func NewRenderer(t Theme) Renderer {
	return Renderer{
		Theme: sanitizeTheme(t),
	}
}

// RenderDocument renders a normalized Document into a theme-aware,
// human-readable string. The result is suitable for terminals, logs,
// and Markdown-capable viewers.
func (r Renderer) RenderDocument(doc *model.Document) string {
	if doc == nil {
		return ""
	}

	var b strings.Builder
	width := r.Theme.EffectiveWidth(80)

	// Title
	title := strings.TrimSpace(doc.Title)
	if title == "" && doc.SourceURL != "" {
		title = strings.TrimSpace(doc.SourceURL)
	}
	if title != "" {
		renderedTitle := r.renderHeading(1, title)
		b.WriteString(renderedTitle)
		b.WriteByte('\n')
		b.WriteByte('\n')
	}

	// Excerpt (if present)
	if strings.TrimSpace(doc.Excerpt) != "" {
		excerpt := wrapTextToWidth(strings.TrimSpace(doc.Excerpt), width)
		excerpt = styleEm(r.Theme, excerpt)
		b.WriteString(excerpt)
		b.WriteByte('\n')
		b.WriteByte('\n')
	}

	// Metadata (compact)
	if len(doc.Metadata) > 0 {
		metaLines := r.renderMetadata(doc.Metadata, width)
		if metaLines != "" {
			b.WriteString(metaLines)
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
	}

	// Content (fallback when there are no sections)
	if strings.TrimSpace(doc.Content) != "" && len(doc.Sections) == 0 {
		body := wrapTextToWidth(strings.TrimSpace(doc.Content), width)
		b.WriteString(body)
		return strings.TrimRight(b.String(), "\n")
	}

	// Sections
	for i, s := range doc.Sections {
		sec := r.renderSection(&s, width)
		if sec == "" {
			continue
		}
		b.WriteString(sec)
		if i < len(doc.Sections)-1 {
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

// RenderMarkdown is a convenience alias for RenderDocument, so that
// public APIs can explicitly express the intention to produce
// Markdown-like text.
func (r Renderer) RenderMarkdown(doc *model.Document) string {
	return r.RenderDocument(doc)
}

//
// ────────────────────────────────────────────────────────────────────────
//                           SECTION RENDERING
// ────────────────────────────────────────────────────────────────────────
//

func (r Renderer) renderSection(s *model.Section, width int) string {
	if s == nil {
		return ""
	}

	var b strings.Builder

	// Optional section role label.
	if r.Theme.ShowSectionRoles {
		label := "[" + string(s.Role) + "]"
		b.WriteString(styleMeta(r.Theme, label))
		b.WriteByte('\n')
	}

	heading := strings.TrimSpace(s.Heading)
	text := strings.TrimSpace(s.Text)

	switch s.Role {
	case model.SectionRoleBody, model.SectionRoleSummary, model.SectionRoleUnknown:
		// Article or generic body content.
		if heading != "" {
			h := r.renderHeading(2, heading)
			b.WriteString(h)
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
		if text != "" {
			body := wrapTextToWidth(text, width)
			b.WriteString(body)
		}

	case model.SectionRoleFeedItem:
		// Feed item: heading + text, bullet-style.
		line := heading
		if line == "" {
			line = "(feed item)"
		}
		line = styleStrong(r.Theme, line)
		line = r.Theme.Bullet + " " + line
		line = wrapTextToWidth(line, width)
		b.WriteString(line)

		if text != "" {
			b.WriteByte('\n')
			body := wrapTextToWidth(text, width)
			b.WriteString(body)
		}

	case model.SectionRoleEntity:
		// Entity: heading as strong, summary body, plus metadata.
		if heading != "" {
			h := styleStrong(r.Theme, heading)
			h = wrapTextToWidth(h, width)
			b.WriteString(h)
			b.WriteByte('\n')
		}
		if text != "" {
			body := wrapTextToWidth(text, width)
			b.WriteString(body)
		}
		if len(s.Meta) > 0 {
			if text != "" {
				b.WriteByte('\n')
			}
			metaLines := r.renderMetadata(s.Meta, width)
			if metaLines != "" {
				b.WriteString(metaLines)
			}
		}

	case model.SectionRoleMetadata:
		// Pure metadata section.
		if heading != "" {
			h := r.renderHeading(3, heading)
			b.WriteString(h)
			b.WriteByte('\n')
		}
		if len(s.Meta) > 0 {
			metaLines := r.renderMetadata(s.Meta, width)
			if metaLines != "" {
				b.WriteString(metaLines)
			}
		}

	default:
		// Fallback: body-like rendering.
		if heading != "" {
			h := r.renderHeading(2, heading)
			b.WriteString(h)
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
		if text != "" {
			body := wrapTextToWidth(text, width)
			b.WriteString(body)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

//
// ────────────────────────────────────────────────────────────────────────
//                            METADATA RENDERING
// ────────────────────────────────────────────────────────────────────────
//

func (r Renderer) renderMetadata(meta map[string]string, width int) string {
	if len(meta) == 0 {
		return ""
	}

	var b strings.Builder
	for k, v := range meta {
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)
		if key == "" || val == "" {
			continue
		}

		line := key + ": " + val
		line = wrapTextToWidth(line, width)
		line = styleMeta(r.Theme, line)

		b.WriteString(line)
		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}

//
// ────────────────────────────────────────────────────────────────────────
//                          HEADING RENDERING
// ────────────────────────────────────────────────────────────────────────
//

func (r Renderer) renderHeading(level int, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	style := r.Theme.HeadingForLevel(level)

	headingText := text
	if style.Uppercase {
		headingText = strings.ToUpper(headingText)
	}

	var b strings.Builder

	// Prefix-based style (e.g. "# ", "## ", "• ").
	if style.Prefix != "" {
		b.WriteString(style.Prefix)
		b.WriteString(headingText)
	} else {
		b.WriteString(headingText)
	}

	rendered := b.String()

	// Apply ANSI styling for headings.
	rendered = styleHeading(r.Theme, rendered)

	// Optional underline style:
	//   Title
	//   =====
	if style.Underline {
		underlineRune := style.UnderlineRune
		if underlineRune == 0 {
			underlineRune = '='
		}
		underline := strings.Repeat(string(underlineRune), len(headingText))
		return rendered + "\n" + underline
	}

	return rendered
}
