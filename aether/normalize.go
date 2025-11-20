// aether/normalize.go
//
// Public normalization wrapper for Aether.

package aether

import (
	"encoding/json"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/normalize"
	"github.com/Nibir1/Aether/internal/toon"
)

// Alias for public use.
type NormalizedDocument = model.Document

func (c *Client) NormalizeSearchResult(sr *SearchResult) *NormalizedDocument {
	if c == nil {
		return &model.Document{
			Kind:     model.DocumentKindUnknown,
			Metadata: map[string]string{},
		}
	}
	return normalize.Pipeline(convertSearchResult(sr))
}

// JSON output
func (c *Client) MarshalSearchResultJSON(sr *SearchResult) ([]byte, error) {
	doc := c.NormalizeSearchResult(sr)
	return json.MarshalIndent(doc, "", "  ")
}

// TOON output
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

		// No entities here because aether.SearchResult has none.
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
