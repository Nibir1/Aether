// Package display provides Aether’s rich output-rendering subsystem.
//
// This package is responsible for converting normalized Aether documents
// into human-readable, theme-aware representations such as:
//
//   - Markdown (primary output format)
//   - ANSI-colored terminal views
//   - Width-aware text tables
//   - Short previews for CLI / TUI interfaces
//
// The Display subsystem supports four major features:
//
//  1. ANSI-aware color rendering
//     • Safe fallback when color is not supported
//     • Works in terminals, logs, and non-TTY environments
//
//  2. Theme profiles (Theme)
//     • DefaultTheme
//     • DarkTheme
//     • MinimalTheme
//     • PaperTheme (no ANSI, printer-style)
//
//  3. Adaptive width rendering
//     • Detects terminal width at runtime
//     • Falls back to 80–100 characters when unknown
//     • Provides wrapping, truncation, and preview helpers
//
//  4. Markdown transformation
//     • Heading styles
//     • Code blocks
//     • Lists, links, inline emphasis
//     • Table rendering
//
// Architecture:
//
//	theme.go         → theme definitions + runtime feature detection
//	color.go         → ANSI color & style helpers (with fallbacks)
//	width.go         → terminal width detection + text wrapping
//	markdown.go      → generic Markdown formatting
//	model_render.go  → render model.Document into Markdown
//	table.go         → flexible Unicode/ASCII table renderer
//	preview.go       → short previews (title + excerpt)
//
// The package is intentionally decoupled from:
//   - async fetcher
//   - plugins
//   - normalization
//
// allowing it to be used not only for Aether.Search(), but also for
// rendering OpenAPI results, feed items, or user-generated documents.
//
// Export policy:
//
//	Internal-only. Public rendering APIs live in aether/display.go in later stages.
//
// Stage 18 establishes the foundation for fully themeable, terminal-aware
// output rendering for both human users and LLM pipelines.
package display
