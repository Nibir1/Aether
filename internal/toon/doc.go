// internal/toon/doc.go
//
// Package toon defines Aether's token-oriented normalized document format.
//
// TOON is a loss-controlled, LLM-friendly representation of model.Document.
// Unlike Markdown or plain JSON, TOON expresses a document as a sequence of
// structured tokens that preserve:
//   - document kind (article, feed, html, entity, etc.)
//   - section boundaries and roles
//   - metadata
//   - text, headings, summaries
//
// Goals:
//   - Stable across Aether releases
//   - Easy for LLMs to ingest
//   - Streamable, chunkable, and truncatable
//   - Round-trippable (FromModel → ToModel best-effort)
//
// Example TOON token:
//
//	{
//	  "type": "section_start",
//	  "role": "body",
//	  "attrs": {"heading": "Introduction"}
//	}
//
// TOON 2.0 introduces:
//   - explicit token types
//   - section boundaries
//   - document-level metadata
//   - role-aware segmentation
package toon
