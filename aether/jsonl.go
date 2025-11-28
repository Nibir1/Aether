// aether/jsonl.go
//
// JSONL Streaming Output for Aether
//
// This file introduces a streaming interface that writes normalized
// Aether documents as *JSON Lines* (JSONL), one logical unit per line,
// enabling incremental consumption by CLIs, pipelines, and LLM/RAG systems.
//
// Design Rules:
//   • Streaming is incremental (no buffering)
//   • Each line is a valid standalone JSON object
//   • Output units follow: Document → Metadata → Sections → FeedItems
//   • Caller controls the io.Writer (stdout, file, socket, pipe)
//   • Zero memory accumulation (ideal for large documents)
//   • NO plugins are executed here — this is purely serialization.
//
// The goal is to provide a production-grade JSONL interface without
// altering core normalization or TOON logic.

package aether

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Nibir1/Aether/internal/rss"
)

// JSONLObject is the container for each JSONL line.
// Every streamed line uses this structure.
type JSONLObject struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// writeJSONL writes a single JSONL object to w.
func writeJSONL(w io.Writer, obj JSONLObject) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

//
// ────────────────────────────────────────────────────────────────
//           PUBLIC API — STREAM NORMALIZED DOCUMENT
// ────────────────────────────────────────────────────────────────
//

// StreamNormalizedJSONL streams a normalized Document as JSONL.
func (c *Client) StreamNormalizedJSONL(ctx context.Context, w io.Writer, doc *NormalizedDocument) error {
	if doc == nil {
		return fmt.Errorf("aether: nil document in StreamNormalizedJSONL")
	}

	// 1) Stream the document header
	if err := writeJSONL(w, JSONLObject{
		Type: "document",
		Data: map[string]interface{}{
			"kind":       doc.Kind,
			"title":      doc.Title,
			"excerpt":    doc.Excerpt,
			"content":    doc.Content,
			"source_url": doc.SourceURL,
		},
	}); err != nil {
		return err
	}

	// 2) Stream metadata
	if len(doc.Metadata) > 0 {
		if err := writeJSONL(w, JSONLObject{
			Type: "metadata",
			Data: map[string]interface{}{
				"metadata": doc.Metadata,
			},
		}); err != nil {
			return err
		}
	}

	// 3) Stream sections one-by-one
	for _, s := range doc.Sections {
		if err := writeJSONL(w, JSONLObject{
			Type: "section",
			Data: map[string]interface{}{
				"role":    s.Role,
				"heading": s.Heading,
				"text":    s.Text,
				"meta":    s.Meta,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

//
// ────────────────────────────────────────────────────────────────
//       PUBLIC API — STREAM SEARCHRESULT DIRECTLY AS JSONL
// ────────────────────────────────────────────────────────────────
//

// StreamSearchResultJSONL normalizes a SearchResult and streams JSONL.
func (c *Client) StreamSearchResultJSONL(ctx context.Context, w io.Writer, sr *SearchResult) error {
	if sr == nil {
		return fmt.Errorf("aether: nil SearchResult")
	}
	doc := c.NormalizeSearchResult(sr)
	return c.StreamNormalizedJSONL(ctx, w, doc)
}

//
// ────────────────────────────────────────────────────────────────
//     OPTIONAL: STREAM FEED ITEMS DIRECTLY (RSS / Atom / etc.)
// ────────────────────────────────────────────────────────────────
//

// StreamFeedJSONL streams each FeedItem as a JSONL object.
func (c *Client) StreamFeedJSONL(ctx context.Context, w io.Writer, feed *rss.Feed) error {
	if feed == nil {
		return fmt.Errorf("aether: nil Feed")
	}

	for _, item := range feed.Items {
		obj := JSONLObject{
			Type: "feed_item",
			Data: map[string]interface{}{
				"title":       item.Title,
				"link":        item.Link,
				"author":      item.Author,
				"guid":        item.GUID,
				"description": item.Description,
				"content":     item.Content,
				"published":   item.Published,
				"updated":     item.Updated,
			},
		}
		if err := writeJSONL(w, obj); err != nil {
			return err
		}
	}
	return nil
}
