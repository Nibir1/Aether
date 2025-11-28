// aether/toon_stream.go
//
// TOON streaming interface (JSONL event stream).
//
// This file provides a streaming representation of Aether's TOON 2.0
// documents. Instead of producing one large TOON JSON, Aether emits
// a JSONL event sequence:
//
//   doc_start → doc_meta → token* → doc_end
//
// This enables:
//   • streaming to LLM/RAG pipelines
//   • incremental indexing
//   • CLI streaming consumption (jq, awk, etc.)
//   • message-queue style processing
//
// All writes respect context cancellation, and no entire document is
// ever buffered in memory.

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
//                        PUBLIC ENTRYPOINTS
// ───────────────────────────────────────────────────────────────────────────
//

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

	tdoc := c.ToTOONFromModel(doc)
	return streamTOONDocument(ctx, w, tdoc)
}

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
//                           INTERNAL STREAMING
// ───────────────────────────────────────────────────────────────────────────
//

func streamTOONDocument(ctx context.Context, w io.Writer, doc *toon.Document) error {
	if doc == nil {
		doc = &toon.Document{}
	}

	enc := json.NewEncoder(w)

	// 1. doc_start
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

	// 2. doc_meta
	if len(doc.Attributes) > 0 {
		metaEv := toonStreamEvent{
			Event: "doc_meta",

			// IMPORTANT:
			// doc.Attributes is map[string]string → so Attrs must be map[string]string
			Attrs: doc.Attributes,
		}
		if err := encodeTOONEvent(ctx, enc, &metaEv); err != nil {
			return err
		}
	}

	// 3. token events
	for _, tok := range doc.Tokens {
		ev := toonStreamEvent{
			Event: "token",
			Token: &toonStreamToken{
				Type:     string(tok.Type),
				Category: categorizeTOONToken(tok),
				Role:     tok.Role,
				Text:     tok.Text,

				// tok.Attrs is also map[string]string in your model
				Attrs: tok.Attrs,
			},
		}

		if err := encodeTOONEvent(ctx, enc, &ev); err != nil {
			return err
		}
	}

	// 4. doc_end
	end := toonStreamEvent{Event: "doc_end"}
	return encodeTOONEvent(ctx, enc, &end)
}

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
//                           TOKEN CATEGORY LOGIC
// ───────────────────────────────────────────────────────────────────────────
//

func categorizeTOONToken(t toon.Token) string {
	switch {
	case t.IsSectionBoundary():
		return "boundary"
	case t.IsContentToken():
		return "content"
	case t.IsMetadata():
		return "metadata"
	default:
		return "other"
	}
}

//
// ───────────────────────────────────────────────────────────────────────────
//                           STREAM EVENT STRUCTS
// ───────────────────────────────────────────────────────────────────────────
//

type toonStreamEvent struct {
	Event   string `json:"event"`
	Kind    string `json:"kind,omitempty"`
	Source  string `json:"source_url,omitempty"`
	Title   string `json:"title,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`

	// FIXED: must match toon.Attributes type = map[string]string
	Attrs map[string]string `json:"attrs,omitempty"`

	Token *toonStreamToken `json:"token,omitempty"`
}

type toonStreamToken struct {
	Type     string `json:"type"`
	Category string `json:"category"`
	Role     string `json:"role,omitempty"`
	Text     string `json:"text,omitempty"`

	// FIXED: must match toon.Token.Attrs type = map[string]string
	Attrs map[string]string `json:"attrs,omitempty"`
}
