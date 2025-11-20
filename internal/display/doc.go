// internal/display/doc.go
//
// Package display implements Aether’s presentation layer. It provides a
// theme–aware Markdown rendering pipeline for normalized documents
// (internal/model.Document) produced by the Normalize subsystem.
//
// The Display subsystem is intentionally modular and is split across several
// focused source files:
//
//	markdown.go      – Markdown renderer (primary user-facing renderer)
//	themes.go        – Built-in Markdown theme definitions
//	model_render.go  – Internal helpers for turning model.Document content
//	                    into render-ready components (paragraphs, sections,
//	                    metadata blocks). Designed for future non-Markdown
//	                    output formats.
//	preview.go       – Document preview generator (short excerpts, summaries,
//	                    feed overviews). Used in later Search and TUI layers.
//	table.go         – Markdown table rendering utilities for normalized
//	                    metadata and structured content
//
// # Rendering Philosophy
//
// Aether is designed to be LLM-friendly and human-readable. The Display
// subsystem therefore follows these guiding principles:
//
//   - Pure, standard Markdown — no HTML, no ANSI escape codes.
//     (ANSI colors will be introduced in a later dedicated stage.)
//
//   - Themeable output. All rendered Markdown can use one of several
//     built-in themes (GitHub Dark, Solarized, Gruvbox, Minimalist).
//     Themes alter visual accents while remaining fully valid Markdown.
//
//   - Deterministic output. The same input document will always render
//     identically under a given theme, ensuring stable hashing,
//     deterministic caching, and reproducible LLM prompts.
//
//   - Structure-preserving. Sections, metadata, paragraphs, summaries, and
//     feed items are rendered in a way that preserves semantic meaning,
//     enabling better downstream reasoning by models.
//
// # Input Model
//
// The renderer consumes the normalized internal/model.Document type.
// Every SearchResult eventually becomes a normalized Document, meaning the
// Display subsystem does not need to understand:
//   - HTML trees
//   - extracted article metadata
//   - feed formats
//   - raw HTTP response bodies
//
// All extraction and normalization happens earlier in the pipeline.
//
// # Extensibility
//
// While Stage 12 introduces only Markdown rendering, the architecture is
// intentionally designed to support:
//
//   - plain-text output
//   - ANSI-styled CLI output
//   - HTML rendering
//   - TOON previews
//   - TUI widgets (tables, cards, summaries)
//
// These will be added in future stages without needing to redesign the
// Display subsystem.
//
// # Usage (public API)
//
// Aether's public package exposes:
//
//	client.RenderMarkdown(doc)
//	client.RenderMarkdownWithTheme(doc, theme)
//
// where `doc` is a *normalized* document from NormalizeSearchResult.
//
// This file contains no logic. It exists solely to document the package.
package display
