// internal/html/util.go
//
// Utility helpers for text normalization and whitespace handling.

package html

import "strings"

// cleanWhitespace collapses runs of whitespace into a single space and trims.
func cleanWhitespace(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Replace all whitespace sequences with a single space.
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false

	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return b.String()
}
