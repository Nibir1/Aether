// aether/toon_stream.go
//
// TOON streaming interface (JSONL event stream).
//
// This file provides a streaming representation of Aether's TOON 2.0
// documents. Instead of emitting one large TOON JSON object, Aether
// writes a sequence of JSON lines, each describing a small "event":
//
//   • doc_start  — top-level document header (kind, URL, title, excerpt)
//   • doc_meta   — document-level attributes (TOON Attributes map)
//   • token      — individual TOON tokens (text, headings, sections, meta)
//   • doc_end    — end-of-document marker
//
// The stream is line-delimited JSON (JSONL), making it easy to consume:
//
//   • CLI tools (e.g., jq, awk, streaming processors)
//   • LLM/RAG ingestion pipelines
//   • message queues and event-stream systems
//
// Tokens are annotated with a coarse "category":
//   - "boundary"  → section_start / section_end
//   - "content"   → text / heading / title / excerpt
//   - "metadata"  → meta / docinfo
//   - "other"     → any future token kinds
//
// This hybrid design (Option C) keeps the token-level fidelity of TOON
// while making the stream semantically rich and easy to index.
//
// All functions here are safe for use with large documents and respect
// context cancellation between writes.

package aether

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Nibir1/Aether/internal/toon"
)

//
// ───────────────────────────────────────────────────────────────────────────
//                        PUBLIC STREAMING ENTRYPOINTS
// ───────────────────────────────────────────────────────────────────────────
//

// StreamTOON streams a normalized document as a TOON event sequence in JSONL.
//
//   - ctx controls cancellation
//   - w is the destination writer (stdout, file, socket, etc.)
//   - doc is a normalized Aether document (model.Document / NormalizedDocument)
//
// The function converts the document to TOON 2.0 and then emits:
//
//	doc_start → doc_meta? → token* → doc_end
func (c *Client) StreamTOON(ctx context.Context, w io.Writer, doc *NormalizedDocument) error {
	if c == nil {
		return fmt.Errorf("aether: nil client in StreamTOON")
	}
	if w == nil {
		return fmt.Errorf("aether: nil writer in StreamTOON")
	}
	if doc == nil {
		return fmt.Errorf("aether: nil document in StreamTOON")
	}

	// doc is already a *model.Document because NormalizedDocument = model.Document
	tdoc := c.ToTOONFromModel(doc)
	return streamTOONDocument(ctx, w, tdoc)
}

// StreamSearchResultTOON normalizes a SearchResult and streams it as
// TOON JSONL events using the same event schema as StreamTOON.
func (c *Client) StreamSearchResultTOON(ctx context.Context, w io.Writer, sr *SearchResult) error {
	if c == nil {
		return fmt.Errorf("aether: nil client in StreamSearchResultTOON")
	}
	if w == nil {
		return fmt.Errorf("aether: nil writer in StreamSearchResultTOON")
	}
	if sr == nil {
		return fmt.Errorf("aether: nil SearchResult in StreamSearchResultTOON")
	}

	tdoc := c.ToTOON(sr)
	return streamTOONDocument(ctx, w, tdoc)
}

//
// ───────────────────────────────────────────────────────────────────────────
//                          INTERNAL STREAMING LOGIC
// ───────────────────────────────────────────────────────────────────────────
//

// streamTOONDocument emits the full TOON event sequence for a single document.
func streamTOONDocument(ctx context.Context, w io.Writer, doc *toon.Document) error {
	if doc == nil {
		doc = &toon.Document{}
	}

	enc := json.NewEncoder(w)

	// 1) doc_start
	start := toonStreamEvent{
		Event:   "doc_start",
		Kind:    string(doc.Kind),
		Source:  doc.SourceURL,
		Title:   doc.Title,
		Excerpt: doc.Excerpt,
	}
	if err := encodeTOONEvent(ctx, enc, &start); err != nil {
		return err
	}

	// 2) doc_meta (if any attributes present)
	if len(doc.Attributes) > 0 {
		metaEv := toonStreamEvent{
			Event: "doc_meta",
			Attrs: doc.Attributes,
		}
		if err := encodeTOONEvent(ctx, enc, &metaEv); err != nil {
			return err
		}
	}

	// 3) token events
	for _, tok := range doc.Tokens {
		ev := toonStreamEvent{
			Event: "token",
			Token: &toonStreamToken{
				Type:     string(tok.Type),
				Category: categorizeTOONToken(tok),
				Role:     tok.Role,
				Text:     tok.Text,
				Attrs:    tok.Attrs,
			},
		}

		if err := encodeTOONEvent(ctx, enc, &ev); err != nil {
			return err
		}
	}

	// 4) doc_end
	end := toonStreamEvent{
		Event: "doc_end",
	}
	if err := encodeTOONEvent(ctx, enc, &end); err != nil {
		return err
	}

	return nil
}

// encodeTOONEvent encodes a single event as one JSON line,
// respecting context cancellation.
func encodeTOONEvent(ctx context.Context, enc *json.Encoder, ev *toonStreamEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return enc.Encode(ev)
	}
}

//
// ───────────────────────────────────────────────────────────────────────────
//                           TOKEN CATEGORIZATION
// ───────────────────────────────────────────────────────────────────────────
//

// categorizeTOONToken maps a raw TOON token into a coarse-grained category
// used in the streaming representation for easier downstream routing.
//
// Categories:
//   - "boundary"  → section_start / section_end
//   - "content"   → text, heading, title, excerpt
//   - "metadata"  → docinfo, meta
//   - "other"     → anything else (forward-compatible)
func categorizeTOONToken(t toon.Token) string {
	if t.IsSectionBoundary() {
		return "boundary"
	}
	if t.IsContentToken() {
		return "content"
	}
	if t.IsMetadata() {
		return "metadata"
	}
	return "other"
}

//
// ───────────────────────────────────────────────────────────────────────────
//                           STREAM EVENT STRUCTURES
// ───────────────────────────────────────────────────────────────────────────
//

// toonStreamEvent represents a single JSONL event in the TOON stream.
//
// Event values:
//   - "doc_start" — top-level document header
//   - "doc_meta"  — document-level attributes
//   - "token"     — individual TOON token
//   - "doc_end"   — end-of-document marker
type toonStreamEvent struct {
	Event   string            `json:"event"`          // doc_start, doc_meta, token, doc_end
	Kind    string            `json:"kind,omitempty"` // document kind
	Source  string            `json:"source_url,omitempty"`
	Title   string            `json:"title,omitempty"`
	Excerpt string            `json:"excerpt,omitempty"`
	Attrs   map[string]string `json:"attrs,omitempty"` // document-level TOON attributes
	Token   *toonStreamToken  `json:"token,omitempty"` // token payload
}

// toonStreamToken is the streamed representation of a single TOON token.
type toonStreamToken struct {
	Type     string            `json:"type"`            // e.g., "text", "heading", ...
	Category string            `json:"category"`        // boundary | content | metadata | other
	Role     string            `json:"role,omitempty"`  // semantic role
	Text     string            `json:"text,omitempty"`  // text content
	Attrs    map[string]string `json:"attrs,omitempty"` // token attributes
}
