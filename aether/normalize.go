// aether/normalize.go
//
// This file exposes Aether's normalization utilities for converting
// high-level SearchResult values into canonical normalized documents
// and then serializing them as plain JSON or TOON.
//
// JSON output uses the internal model.Document schema.
// TOON output uses the token-oriented internal/toon.Document schema.

package aether

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

// NormalizedDocument is the public alias for the internal normalized
// document model. Callers that want to inspect the normalized structure
// directly (in Go) can use this type.
type NormalizedDocument = model.Document

// NormalizeSearchResult converts a SearchResult into a normalized
// Document.
//
// It focuses on the PrimaryDocument and, when present, incorporates
// Article / Feed views into sections and metadata.
func (c *Client) NormalizeSearchResult(sr *SearchResult) *NormalizedDocument {
	if sr == nil || sr.PrimaryDocument == nil {
		return &model.Document{
			Kind:      model.DocumentKindUnknown,
			Metadata:  map[string]string{},
			Sections:  nil,
			Content:   "",
			Title:     "",
			Excerpt:   "",
			SourceURL: "",
		}
	}

	doc := sr.PrimaryDocument

	var kind model.DocumentKind
	switch doc.Kind {
	case SearchDocumentKindArticle:
		kind = model.DocumentKindArticle
	case SearchDocumentKindHTML:
		kind = model.DocumentKindHTML
	case SearchDocumentKindFeed:
		kind = model.DocumentKindFeed
	case SearchDocumentKindJSON:
		kind = model.DocumentKindJSON
	case SearchDocumentKindText:
		kind = model.DocumentKindText
	case SearchDocumentKindBinary:
		kind = model.DocumentKindBinary
	default:
		kind = model.DocumentKindUnknown
	}

	meta := map[string]string{}
	for k, v := range doc.Metadata {
		meta[k] = v
	}
	// Include search plan intent as metadata if available.
	if sr.Plan.Intent != "" {
		meta["aether.intent"] = string(sr.Plan.Intent)
	}

	normalized := &model.Document{
		SourceURL: doc.URL,
		Kind:      kind,
		Title:     doc.Title,
		Excerpt:   doc.Excerpt,
		Content:   strings.TrimSpace(doc.Content),
		Metadata:  meta,
		Sections:  nil,
	}

	// Incorporate Article details into a dedicated section, if available.
	if sr.Article != nil && strings.TrimSpace(sr.Article.Content) != "" {
		section := model.Section{
			Role:    model.SectionRoleBody,
			Heading: sr.Article.Title,
			Text:    sr.Article.Content,
			Meta:    sr.Article.Meta,
		}
		normalized.Sections = append(normalized.Sections, section)
	}

	// Incorporate Feed items as sections when a feed is present.
	if sr.Feed != nil && len(sr.Feed.Items) > 0 {
		for _, item := range sr.Feed.Items {
			body := chooseFirstNonEmpty(
				item.Content,
				item.Description,
				item.Title,
			)

			section := model.Section{
				Role:    model.SectionRoleFeedItem,
				Heading: item.Title,
				Text:    body,
				Meta: map[string]string{
					"link":           item.Link,
					"author":         item.Author,
					"guid":           item.GUID,
					"published_unix": int64ToString(item.Published),
					"updated_unix":   int64ToString(item.Updated),
				},
			}

			normalized.Sections = append(normalized.Sections, section)
		}
	}

	return normalized
}

// MarshalSearchResultJSON serializes a SearchResult into normalized
// JSON using Aether's canonical model.Document schema.
func (c *Client) MarshalSearchResultJSON(sr *SearchResult) ([]byte, error) {
	doc := c.NormalizeSearchResult(sr)
	return json.MarshalIndent(doc, "", "  ")
}

// MarshalSearchResultTOON serializes a SearchResult into TOON format.
func (c *Client) MarshalSearchResultTOON(sr *SearchResult) ([]byte, error) {
	doc := c.NormalizeSearchResult(sr)
	tdoc := toon.FromModel(doc)
	return json.MarshalIndent(tdoc, "", "  ")
}

// chooseFirstNonEmpty returns the first non-empty string from the list.
func chooseFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// int64ToString converts an int64 timestamp to string.
func int64ToString(ts int64) string {
	if ts == 0 {
		return ""
	}
	return strconv.FormatInt(ts, 10)
}
