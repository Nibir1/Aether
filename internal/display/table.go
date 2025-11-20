// internal/display/table.go
//
// This file implements Aether’s Markdown table renderer.
// It provides utilities for rendering structured tabular data
// such as metadata maps, OpenAPI responses, feed lists, or
// any structured key/value data.
//
// The resulting tables follow GitHub Flavored Markdown (GFM) rules.
// All output is deterministic, consistent, and pure Markdown
// (no HTML, no ANSI).
//
// Features:
//   - Automatic column-width calculation
//   - Proper escaping of Markdown-breaking characters
//   - Deterministic column ordering for maps (sorted by key)
//   - LLM-friendly formatting
//   - No dependencies, pure Go.

package display

import (
	"sort"
	"strings"
)

// Table represents a generic Markdown table.
type Table struct {
	Headers []string
	Rows    [][]string
}

// RenderTable returns a Markdown-formatted table.
// The returned string does NOT include surrounding blank lines, giving
// the caller control over spacing.
//
// Rules:
//   - Headers define the number of columns.
//   - All rows must have the same number of columns.
//   - Markdown characters are properly escaped.
//   - Columns auto-size to the widest cell.
func RenderTable(t Table) string {
	if len(t.Headers) == 0 {
		return ""
	}

	cols := len(t.Headers)
	widths := make([]int, cols)

	// Calculate column widths from headers first.
	for i, h := range t.Headers {
		w := len(escapeTableContent(h))
		if w > widths[i] {
			widths[i] = w
		}
	}

	// Now size from row data.
	for _, row := range t.Rows {
		if len(row) != cols {
			// Skip malformed rows silently; Aether does
			// not allow panics in rendering.
			continue
		}
		for i, cell := range row {
			w := len(escapeTableContent(cell))
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	var b strings.Builder

	// Write header row.
	for i, h := range t.Headers {
		b.WriteString("| ")
		b.WriteString(padRight(escapeTableContent(h), widths[i]))
		b.WriteString(" ")
	}
	b.WriteString("|\n")

	// Write separator row.
	for i := range t.Headers {
		b.WriteString("| ")
		b.WriteString(strings.Repeat("-", widths[i]))
		b.WriteString(" ")
	}
	b.WriteString("|\n")

	// Write data rows.
	for _, row := range t.Rows {
		if len(row) != cols {
			continue
		}
		for i, cell := range row {
			b.WriteString("| ")
			b.WriteString(padRight(escapeTableContent(cell), widths[i]))
			b.WriteString(" ")
		}
		b.WriteString("|\n")
	}

	return b.String()
}

// RenderKeyValueTable renders a map[string]string as a two-column
// Markdown table, sorted by key. This is ideal for metadata blocks
// or flattening JSON-like objects for LLMs or CLI display.
func RenderKeyValueTable(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	// Stable order: sorted keys.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t := Table{
		Headers: []string{"Key", "Value"},
		Rows:    make([][]string, 0, len(m)),
	}

	for _, k := range keys {
		t.Rows = append(t.Rows, []string{k, m[k]})
	}

	return RenderTable(t)
}

// escapeTableContent escapes characters that can break Markdown tables.
// We escape pipes and backslashes; headings and other Markdown formatting
// are NOT escaped here, only table-breaking characters.
func escapeTableContent(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return strings.TrimSpace(s)
}

// padRight pads a string with spaces to a target width.
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
