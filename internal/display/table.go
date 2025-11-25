// internal/display/table.go
//
// Theme-aware table rendering for Aether’s Display subsystem.
//
// This renderer supports:
//   • Unicode or ASCII borders (Theme.AsciiOnly)
//   • Column width calculation based on Theme.TablePadding
//   • Word wrapping according to Theme.EffectiveWidth()
//   • Header and body styling via TableStyle (bold, faint, colorize)
//   • ANSI styling via color.go
//
// Used internally by Renderer.RenderTable(), Renderer.RenderPreview(),
// and model_render.go when a document section requests table formatting.

package display

import (
	"strings"
	"unicode/utf8"
)

//
// ─────────────────────────────────────────────────────────────────────────────
//                                TABLE TYPES
// ─────────────────────────────────────────────────────────────────────────────
//

// Table represents a simple table with optional header row.
type Table struct {
	Header []string
	Rows   [][]string
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                               MAIN RENDERER
// ─────────────────────────────────────────────────────────────────────────────
//

// RenderTable renders a Table using theme-aware formatting.
func RenderTable(t Theme, tbl Table) string {
	t = sanitizeTheme(t)

	if len(tbl.Header) == 0 && len(tbl.Rows) == 0 {
		return ""
	}

	// Determine usable width.
	totalWidth := t.EffectiveWidth(80)
	if totalWidth < 20 {
		totalWidth = 20
	}

	// Build unified list of rows.
	all := [][]string{}
	if len(tbl.Header) > 0 {
		all = append(all, tbl.Header)
	}
	all = append(all, tbl.Rows...)

	// Calculate column widths.
	colWidths := computeColumnWidths(all, totalWidth, t.TablePadding)

	var b strings.Builder

	// Header row.
	if len(tbl.Header) > 0 {
		b.WriteString(renderTableRow(t, tbl.Header, colWidths, t.TableHeaderStyle))
		b.WriteByte('\n')
		b.WriteString(renderTableSeparator(t, colWidths))
		b.WriteByte('\n')
	}

	// Body rows.
	for i, r := range tbl.Rows {
		b.WriteString(renderTableRow(t, r, colWidths, t.TableBodyStyle))
		if i < len(tbl.Rows)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                       COLUMN WIDTH CALCULATION
// ─────────────────────────────────────────────────────────────────────────────
//

// computeColumnWidths determines each column’s width with padding and scaling.
func computeColumnWidths(rows [][]string, totalWidth int, pad int) []int {
	if len(rows) == 0 {
		return nil
	}

	cols := len(rows[0])
	w := make([]int, cols)

	// Longest cell in each column.
	for _, row := range rows {
		for c := 0; c < cols; c++ {
			cell := ""
			if c < len(row) {
				cell = row[c]
			}
			n := displayLen(cell)
			if n > w[c] {
				w[c] = n
			}
		}
	}

	// Apply padding.
	sum := 0
	for i := range w {
		w[i] += pad * 2
		sum += w[i]
	}

	// Scale down if exceeding totalWidth.
	if sum > totalWidth {
		scale := float64(totalWidth) / float64(sum)
		for i := range w {
			newW := int(float64(w[i]) * scale)
			if newW < 5 {
				newW = 5
			}
			w[i] = newW
		}
	}

	return w
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                               ROW RENDERING
// ─────────────────────────────────────────────────────────────────────────────
//

// renderTableRow renders a single row with wrapping and styling.
func renderTableRow(th Theme, row []string, widths []int, style TableStyle) string {
	var lines [][]string

	// Wrap each cell to width.
	for i, colWidth := range widths {
		text := ""
		if i < len(row) {
			text = strings.TrimSpace(row[i])
		}
		wrapped := wrapTextToWidth(text, colWidth)
		lines = append(lines, strings.Split(wrapped, "\n"))
	}

	// Determine row height.
	maxH := 1
	for _, cell := range lines {
		if len(cell) > maxH {
			maxH = len(cell)
		}
	}

	b := strings.Builder{}

	// Build lines top-aligned.
	for line := 0; line < maxH; line++ {
		for c := 0; c < len(widths); c++ {
			cell := ""
			if line < len(lines[c]) {
				cell = lines[c][line]
			}

			// Apply ANSI styling.
			cell = applyTableStyle(th, cell, style)

			// Pad to width.
			cell = padRight(cell, widths[c])
			b.WriteString(cell)
		}
		if line < maxH-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// applyTableStyle applies bold, faint, or color styling based on Theme.
func applyTableStyle(t Theme, s string, st TableStyle) string {
	if s == "" {
		return s
	}

	out := s
	if st.Bold {
		out = styleStrong(t, out)
	}
	if st.Faint {
		out = styleMeta(t, out)
	}
	if st.Colorize {
		out = styleCode(t, out)
	}
	return out
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                            SEPARATOR RENDERING
// ─────────────────────────────────────────────────────────────────────────────
//

// renderTableSeparator renders a horizontal separator.
//
// Unicode example:
//
//	─────┼──────────┼─────
//
// ASCII fallback:
//
//	-----+----------+------
func renderTableSeparator(t Theme, widths []int) string {
	var b strings.Builder

	sep := "─"
	joint := "┼"

	if t.AsciiOnly {
		sep = "-"
		joint = "+"
	}

	for i, w := range widths {
		b.WriteString(strings.Repeat(sep, w))
		if i < len(widths)-1 {
			b.WriteString(joint)
		}
	}

	return b.String()
}

//
// ─────────────────────────────────────────────────────────────────────────────
//                               UTILITY FUNCTIONS
// ─────────────────────────────────────────────────────────────────────────────
//

// displayLen returns printable width ignoring ANSI escape codes.
func displayLen(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

// padRight pads to the right with spaces.
func padRight(s string, width int) string {
	n := displayLen(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// stripANSI removes ANSI codes for width calculation.
func stripANSI(s string) string {
	var out strings.Builder
	inEsc := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == 0x1b { // ESC
			inEsc = true
			continue
		}

		if inEsc {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEsc = false
			}
			continue
		}

		out.WriteByte(c)
	}

	return out.String()
}
