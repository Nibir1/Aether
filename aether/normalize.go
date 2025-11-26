// aether/normalize.go
//
// Public normalization wrapper for Aether.

package aether

import (
	"context"
	"encoding/json"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/normalize"
	"github.com/Nibir1/Aether/internal/toon"
	"github.com/Nibir1/Aether/plugins"
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

	// (1) Core normalization pipeline
	doc := normalize.Pipeline(convertSearchResult(sr))
	if doc == nil {
		return &model.Document{
			Kind:     model.DocumentKindUnknown,
			Metadata: map[string]string{},
		}
	}

	// (2) Apply TransformPlugins
	finalDoc := c.applyTransformPlugins(doc)

	return finalDoc
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
		Plan: normalize.NormalizePlan{
			Intent: string(in.Plan.Intent),
		},
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
//     ADAPTER: model.Document → plugins.Document
// ─────────────────────────────────────────────
//

func convertModelToPluginDocument(doc *model.Document) *plugins.Document {
	if doc == nil {
		return nil
	}

	// The *canonical* correct URL for plugins is SourceURL.
	out := &plugins.Document{
		Source:   "aether-normalized",
		URL:      doc.SourceURL, // ← FIXED HERE
		Kind:     plugins.DocumentKind(doc.Kind),
		Title:    doc.Title,
		Excerpt:  doc.Excerpt,
		Content:  doc.Content,
		Metadata: cloneStringMap(doc.Metadata),
	}

	// Normalize URL inside metadata for plugin compatibility
	if doc.SourceURL != "" {
		out.Metadata["source_url"] = doc.SourceURL
	}

	// Convert sections
	for _, s := range doc.Sections {
		out.Sections = append(out.Sections, plugins.Section{
			Role:  plugins.SectionRole(s.Role),
			Title: s.Heading,
			Text:  s.Text,
			Meta:  cloneStringMap(s.Meta),
		})
	}

	return out
}

//
// ─────────────────────────────────────────────
//     ADAPTER: plugins.Document → model.Document
// ─────────────────────────────────────────────
//

func convertPluginDocToModel(pdoc *plugins.Document) *model.Document {
	if pdoc == nil {
		return nil
	}

	out := &model.Document{
		Kind:     model.DocumentKind(pdoc.Kind),
		Title:    pdoc.Title,
		Excerpt:  pdoc.Excerpt,
		Content:  pdoc.Content,
		Metadata: cloneStringMap(pdoc.Metadata),
	}

	// Preferred: explicit URL field in plugin Document
	if pdoc.URL != "" {
		out.SourceURL = pdoc.URL
	}

	// Fallback: metadata URL
	if out.SourceURL == "" {
		if metaURL := out.Metadata["url"]; metaURL != "" {
			out.SourceURL = metaURL
		}
	}

	// Normalize fields for consistency
	if out.SourceURL != "" {
		out.Metadata["source_url"] = out.SourceURL
	}

	// Convert sections
	for _, s := range pdoc.Sections {
		out.Sections = append(out.Sections, model.Section{
			Role:    model.SectionRole(s.Role),
			Heading: s.Title,
			Text:    s.Text,
			Meta:    cloneStringMap(s.Meta),
		})
	}

	return out
}

//
// ─────────────────────────────────────────────
//                     HELPERS
// ─────────────────────────────────────────────
//

func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

//
// ─────────────────────────────────────────────
//           TRANSFORM PLUGIN EXECUTION
// ─────────────────────────────────────────────
//

func (c *Client) applyTransformPlugins(doc *model.Document) *model.Document {
	if c == nil || c.plugins == nil || doc == nil {
		return doc
	}

	names := c.plugins.ListTransforms()
	if len(names) == 0 {
		return doc
	}

	current := doc

	for _, name := range names {
		p := c.plugins.GetTransform(name)
		if p == nil {
			continue
		}

		pdoc := convertModelToPluginDocument(current)
		out, err := p.Apply(context.Background(), pdoc)
		if err != nil || out == nil {
			continue
		}

		next := convertPluginDocToModel(out)
		if next != nil {
			current = next
		}
	}

	return current
}
