// internal/display/width.go
//
// Terminal width detection + safe text wrapping utilities for the
// Display subsystem.
//
// This file provides:
//   • DetectTerminalWidth() – best-effort TTY width retrieval
//   • EffectiveWidth(theme) – theme-aware width with fallbacks
//   • wrapTextToWidth() – low-level greedy wrapper (model_render layer)
//
// Normal rules:
//   • If theme.MaxWidth > 0 → always use it
//   • Else if stdout is a TTY → try TIOCGWINSZ
//   • Else fallback to DefaultWidth (80 chars)
//
// This avoids dependencies like "golang.org/x/term" and maintains
// internal-only logic for width-related display concerns.

package display

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// DefaultWidth is used when terminal size cannot be detected.
const DefaultWidth = 80

// winSize mirrors the system struct used by ioctl(TIOCGWINSZ).
type winSize struct {
	Rows uint16
	Cols uint16
	X    uint16
	Y    uint16
}

// DetectTerminalWidth attempts to read terminal width using ioctl.
// Returns (width, ok).
//
// This is a best-effort detection. If detection fails, ok=false.
func DetectTerminalWidth() (int, bool) {
	ws := &winSize{}

	// Use STDOUT for detection.
	fd := os.Stdout.Fd()

	// Only attempt on character devices.
	if !isTerminal(fd) {
		return 0, false
	}

	// Invoke ioctl(TIOCGWINSZ).
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 {
		return 0, false
	}

	if ws.Cols == 0 {
		return 0, false
	}

	return int(ws.Cols), true
}

// isTerminal checks whether the given file descriptor refers to a TTY.
//
// This avoids pulling in x/term but remains cross-platform safe
// (it simply tries fstat and checks the mode).
func isTerminal(fd uintptr) bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Device file?
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// EffectiveWidth decides what width should ultimately be used for
// rendering under the given Theme.
//
// Priority:
//  1. If theme.MaxWidth > 0 → use that
//  2. If terminal width detected → use min(terminalWidth, theme cap)
//  3. Else → DefaultWidth
func EffectiveWidth(t Theme) int {
	// (1) Explicit theme width always wins.
	if t.MaxWidth > 0 {
		return t.MaxWidth
	}

	// (2) Attempt TTY detection
	if w, ok := DetectTerminalWidth(); ok && w > 0 {
		return w
	}

	// (3) Reliable fallback
	return DefaultWidth
}

//
//────────────────────────────────────────────
//         LOW-LEVEL WRAPPING UTILITIES
//────────────────────────────────────────────
//

// wrapTextToWidth performs greedy word wrapping similar to wrapText(),
// but is intended for internal rendering (e.g. model_render.go).
//
// It differs from markdown.go's wrapText by ensuring:
//   - it never trims existing newlines
//   - paragraphs separated by blank lines remain intact
func wrapTextToWidth(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}

	lines := strings.Split(s, "\n")
	out := strings.Builder{}

	for i, line := range lines {
		line = strings.TrimRight(line, " ")

		if line == "" {
			// Preserve blank lines
			out.WriteByte('\n')
			continue
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			out.WriteByte('\n')
			continue
		}

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

		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}

	return out.String()
}
