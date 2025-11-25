// internal/display/color.go
//
// ANSI color and style helpers for Aether's Display subsystem.
//
// This file is intentionally conservative and minimal:
//   • It does not require external dependencies.
//   • It does not attempt to be a full-featured color library.
//   • It provides just enough functionality for themes and model_render
//     to highlight headings, strong text, etc.
//
// Color usage is controlled by:
//   • Theme.Color (ColorModeAuto, ColorModeAlways, ColorModeNever)
//   • Environment variables (e.g. NO_COLOR, TERM)
//
// The display subsystem should always respect these rules so that
// applications embedding Aether can safely use the output in terminals,
// logs, or files without unexpected escape sequences.

package display

import (
	"os"
	"strings"
	"sync"
)

// ansiStyle represents a pair of ANSI escape codes for styling text.
//
// Example:
//
//	ansiStyle{Open: "\x1b[1m", Close: "\x1b[0m"}   // bold
type ansiStyle struct {
	Open  string
	Close string
}

// Predefined basic styles, chosen to be broadly readable.
// These are intentionally minimal; richer palettes can be layered on
// via themes or future extensions.
var (
	ansiBold      = ansiStyle{Open: "\x1b[1m", Close: "\x1b[0m"}
	ansiFaint     = ansiStyle{Open: "\x1b[2m", Close: "\x1b[0m"}
	ansiItalic    = ansiStyle{Open: "\x1b[3m", Close: "\x1b[0m"}
	ansiUnderline = ansiStyle{Open: "\x1b[4m", Close: "\x1b[0m"}

	ansiBlue    = ansiStyle{Open: "\x1b[34m", Close: "\x1b[39m"}
	ansiCyan    = ansiStyle{Open: "\x1b[36m", Close: "\x1b[39m"}
	ansiGreen   = ansiStyle{Open: "\x1b[32m", Close: "\x1b[39m"}
	ansiYellow  = ansiStyle{Open: "\x1b[33m", Close: "\x1b[39m"}
	ansiMagenta = ansiStyle{Open: "\x1b[35m", Close: "\x1b[39m"}
)

// colorSupport encapsulates lazy detection flags for ANSI support.
var (
	colorSupportOnce sync.Once
	colorSupported   bool
)

// detectColorSupport performs a one-time, best-effort detection of
// whether ANSI colors are likely to be supported in the current
// environment.
//
// The logic is deliberately simple:
//
//   - If NO_COLOR is set → no color.
//   - If TERM is empty or "dumb" → no color.
//   - Otherwise → assume color is supported.
//
// Applications that need stricter or richer logic can wrap Display
// outputs and apply their own transformations.
func detectColorSupport() {
	colorSupportOnce.Do(func() {
		// NO_COLOR explicitly disables colors.
		if _, ok := os.LookupEnv("NO_COLOR"); ok {
			colorSupported = false
			return
		}

		term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
		if term == "" || term == "dumb" {
			colorSupported = false
			return
		}

		// Default: assume color is available.
		colorSupported = true
	})
}

// isColorEnabled returns true if color should be used under the given
// Theme and environment conditions.
func isColorEnabled(t Theme) bool {
	switch t.Color {
	case ColorModeNever:
		return false
	case ColorModeAlways:
		return true
	case ColorModeAuto:
		detectColorSupport()
		return colorSupported
	default:
		// Unknown mode → conservative: no color.
		return false
	}
}

// applyStyle applies a given ansiStyle to text if color is enabled
// for the provided Theme. Otherwise, it returns the text unchanged.
//
// This helper is the primary entry point used by model_render and other
// display components when they want to style headings, strong text, etc.
func applyStyle(t Theme, s string, style ansiStyle) string {
	if s == "" {
		return s
	}
	if !isColorEnabled(t) {
		return s
	}
	return style.Open + s + style.Close
}

// styleHeading is a convenience wrapper to style headings. Callers
// may layer this with other formatting (e.g. prefix, underline).
func styleHeading(t Theme, s string) string {
	// Default: bold + blue heading.
	return applyStyle(t, applyStyle(t, s, ansiBold), ansiBlue)
}

// styleStrong emphasizes text (strong/bold).
func styleStrong(t Theme, s string) string {
	return applyStyle(t, s, ansiBold)
}

// styleEm renders emphasized text (italic/faint).
func styleEm(t Theme, s string) string {
	// Italic is not universally supported. Faint is widely supported.
	// We nest italic inside faint for terminals that support both.
	return applyStyle(t, applyStyle(t, s, ansiItalic), ansiFaint)
}

// styleCode renders inline code in a subtle color.
func styleCode(t Theme, s string) string {
	return applyStyle(t, s, ansiCyan)
}

// styleMeta renders metadata labels in a dim style.
func styleMeta(t Theme, s string) string {
	return applyStyle(t, s, ansiFaint)
}
