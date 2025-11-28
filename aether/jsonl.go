// aether/jsonl.go
//
// JSONL Streaming Output for Aether
//
// This file introduces a streaming interface that writes normalized
// Aether documents as JSON Lines (JSONL). It also supports streaming
// public Aether feed items (RSS/Atom) using the public Feed type.

package aether

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// JSONLObject is the structure written for each JSONL line.
type JSONLObject struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// writeJSONL writes exactly one JSONL object to the writer.
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
//           STREAM NORMALIZED DOCUMENT
// ────────────────────────────────────────────────────────────────
//

// StreamNormalizedJSONL streams a normalized Document as JSONL.
func (c *Client) StreamNormalizedJSONL(ctx context.Context, w io.Writer, doc *NormalizedDocument) error {
	if doc == nil {
		return fmt.Errorf("aether: nil document in StreamNormalizedJSONL")
	}

	// Document header line
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

	// Metadata block
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

	// Sections (one per line)
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
//       STREAM SEARCHRESULT DIRECTLY AS JSONL
// ────────────────────────────────────────────────────────────────
//

func (c *Client) StreamSearchResultJSONL(ctx context.Context, w io.Writer, sr *SearchResult) error {
	if sr == nil {
		return fmt.Errorf("aether: nil SearchResult")
	}
	doc := c.NormalizeSearchResult(sr)
	return c.StreamNormalizedJSONL(ctx, w, doc)
}

//
// ────────────────────────────────────────────────────────────────
//       STREAM PUBLIC FEED ITEMS (Aether.Feed)
// ────────────────────────────────────────────────────────────────
//

// StreamFeedJSONL streams each *public* FeedItem as JSONL.
func (c *Client) StreamFeedJSONL(ctx context.Context, w io.Writer, feed *Feed) error {
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
