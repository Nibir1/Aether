// aether/normalize.go
//
// Public normalization wrapper for Aether.
//
// This file adapts the public aether.SearchResult into the internal
// normalize.SearchResult, runs the core normalization pipeline, and then
// applies any registered TransformPlugins over the resulting
// internal/model.Document.
//
// The Model ↔ Plugin adapter logic lives in adapter.go to avoid duplication.

package aether

import (
	"context"
	"encoding/json"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/normalize"
	"github.com/Nibir1/Aether/internal/toon"
)

// Alias for public use.
type NormalizedDocument = model.Document

// NormalizeSearchResult converts a public SearchResult into a canonical
// normalized Document and applies TransformPlugins (if any).
func (c *Client) NormalizeSearchResult(sr *SearchResult) *NormalizedDocument {
	if c == nil {
		return &model.Document{
			Kind:     model.DocumentKindUnknown,
			Metadata: map[string]string{},
		}
	}

	// (1) Core normalization pipeline
	doc := normalize.Pipeline(convertSearchResult(sr))
	if doc == nil {
		return &model.Document{
			Kind:     model.DocumentKindUnknown,
			Metadata: map[string]string{},
		}
	}

	// (2) Apply TransformPlugins (if registered)
	finalDoc := c.applyTransformPlugins(doc)

	return finalDoc
}

// MarshalSearchResultJSON returns pretty-printed JSON for a SearchResult.
func (c *Client) MarshalSearchResultJSON(sr *SearchResult) ([]byte, error) {
	doc := c.NormalizeSearchResult(sr)
	return json.MarshalIndent(doc, "", "  ")
}

// MarshalSearchResultTOON returns pretty-printed TOON JSON for a SearchResult.
func (c *Client) MarshalSearchResultTOON(sr *SearchResult) ([]byte, error) {
	doc := c.NormalizeSearchResult(sr)
	tdoc := toon.FromModel(doc)
	return json.MarshalIndent(tdoc, "", "  ")
}

//
// ─────────────────────────────────────────────
//         ADAPTER: public → internal types
// ─────────────────────────────────────────────
//

// convertSearchResult adapts aether.SearchResult into normalize.SearchResult.
func convertSearchResult(in *SearchResult) *normalize.SearchResult {
	if in == nil {
		return nil
	}

	return &normalize.SearchResult{
		PrimaryDocument: convertPrimaryDocument(in.PrimaryDocument),
		Article:         convertArticle(in.Article),
		Feed:            convertFeed(in.Feed),

		// Only plan intent exists in the public API
		Plan: normalize.NormalizePlan{
			Intent: string(in.Plan.Intent),
		},

		// No entities here because aether.SearchResult has none (yet).
	}
}

func convertPrimaryDocument(in *SearchDocument) *normalize.SearchDocument {
	if in == nil {
		return nil
	}

	return &normalize.SearchDocument{
		URL:      in.URL,
		Title:    in.Title,
		Excerpt:  in.Excerpt,
		Content:  in.Content,
		Metadata: in.Metadata,
		Kind:     string(in.Kind),
	}
}

func convertArticle(in *Article) *normalize.Article {
	if in == nil {
		return nil
	}
	return &normalize.Article{
		Title:   in.Title,
		Content: in.Content,
		Meta:    in.Meta,
	}
}

func convertFeed(in *Feed) *normalize.Feed {
	if in == nil {
		return nil
	}

	out := &normalize.Feed{}
	for _, item := range in.Items {
		out.Items = append(out.Items, normalize.FeedItem{
			Title:       item.Title,
			Link:        item.Link,
			Author:      item.Author,
			GUID:        item.GUID,
			Description: item.Description,
			Content:     item.Content,
			Published:   item.Published,
			Updated:     item.Updated,
		})
	}
	return out
}

//
// ─────────────────────────────────────────────
//         TRANSFORM PLUGIN EXECUTION PIPELINE
// ─────────────────────────────────────────────
//

// applyTransformPlugins runs registered TransformPlugins in registration order.
//
// It converts model.Document ↔ plugins.Document using the unified adapter
// helpers in adapter.go. Failures in individual plugins are logged by the
// caller (in future) but do not abort the pipeline: a failing plugin is
// simply skipped.
func (c *Client) applyTransformPlugins(doc *model.Document) *model.Document {
	if c == nil || c.plugins == nil || doc == nil {
		return doc
	}

	names := c.plugins.ListTransforms()
	if len(names) == 0 {
		return doc // no transforms registered
	}

	current := doc

	for _, name := range names {
		p := c.plugins.GetTransform(name)
		if p == nil {
			continue
		}

		// Convert model.Document → plugins.Document
		pdoc := modelToPluginDocument(current)

		// Execute transform plugin
		out, err := p.Apply(context.Background(), pdoc)
		if err != nil || out == nil {
			// Skip failing plugin but keep the rest of the pipeline.
			continue
		}

		// Convert plugins.Document → model.Document
		next := pluginToModelDocument(out)
		if next != nil {
			current = next
		}
	}

	return current
}
