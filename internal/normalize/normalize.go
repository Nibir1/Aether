// internal/normalize/normalize.go
//
// Orchestrator for Aether's normalization pipeline.
//
// This file provides the main entry points for transforming heterogeneous,
// source-specific data structures (SearchResult, Article, Feed, Entities)
// into the canonical model.Document schema used by Aether.
//
// The pipeline design is intentionally layered:
//
//   1. preNormalize()   – future hook for plugin-based preprocessing
//   2. coreNormalize()  – schema-specific normalization (search, article, feed...)
//   3. merge()          – merge partial Documents (handled in merge.go)
//   4. postNormalize()  – future hook for plugin-based postprocessing
//
// By keeping the orchestrator small and delegating schema-specific logic
// to schema_* files, Aether maintains clarity and extensibility.

package normalize

import (
	"github.com/Nibir1/Aether/internal/model"
)

// SearchResult is the minimal subset of aether.SearchResult that the
// normalization pipeline needs. We define a tiny internal mirror here
// so the normalize package does not depend on the full aether API.
type SearchResult struct {
	PrimaryDocument *SearchDocument
	Article         *Article
	Feed            *Feed
	Entities        []*Entity
	Plan            NormalizePlan
}

// NormalizePlan records meta-intent for the search pipeline (e.g., "lookup",
// "news", "article"), which is incorporated into metadata.
type NormalizePlan struct {
	Intent string
}

// SearchDocument represents the extracted or fetched source document
// from Aether.Search. It mirrors the structure used in aether/search.go.
type SearchDocument struct {
	URL      string
	Title    string
	Excerpt  string
	Content  string
	Metadata map[string]string
	Kind     string // e.g., "article", "feed", "html", …
}

// Article is the extracted readability-based article.
type Article struct {
	Title   string
	Content string
	Meta    map[string]string
}

// Feed is the normalized RSS/Atom representation.
type Feed struct {
	Items []FeedItem
}

type FeedItem struct {
	Title       string
	Link        string
	Author      string
	GUID        string
	Description string
	Content     string
	Published   int64
	Updated     int64
}

// Entity represents a structured API response (Wikidata, etc.).
type Entity struct {
	ID       string
	Label    string
	Summary  string
	URL      string
	Metadata map[string]string
}

// Pipeline orchestrates top-level normalization.
//
// sr may include SearchDocument, Article extraction, Feed data, and
// structured Entities. Each schema_* normalizer can emit a partial
// model.Document. mergeDocuments() combines them into one canonical
// Document.
func Pipeline(sr *SearchResult) *model.Document {
	if sr == nil {
		return emptyDocument()
	}

	partials := []*model.Document{}

	// Future hook:
	sr = preNormalize(sr)

	// Core normalizers:
	if sr.PrimaryDocument != nil {
		partials = append(partials, normalizeSearchDocument(sr))
	}
	if sr.Article != nil {
		partials = append(partials, normalizeArticle(sr))
	}
	if sr.Feed != nil {
		partials = append(partials, normalizeFeed(sr))
	}
	if len(sr.Entities) > 0 {
		partials = append(partials, normalizeEntities(sr))
	}

	// Merge into a single canonical model.Document.
	doc := mergeDocuments(partials...)

	// Add search plan intent, if any.
	if sr.Plan.Intent != "" {
		if doc.Metadata == nil {
			doc.Metadata = map[string]string{}
		}
		doc.Metadata["aether.intent"] = sr.Plan.Intent
	}

	// Future hook:
	doc = postNormalize(doc)

	return doc
}

//
// ────────────────────────────────────────────────────────────────────────
//                   PRE / POST NORMALIZATION HOOKS
// ────────────────────────────────────────────────────────────────────────
//

// preNormalize is a placeholder for future plugin-based preprocessing.
// It simply returns the input unchanged at this stage.
func preNormalize(sr *SearchResult) *SearchResult {
	return sr
}

// postNormalize is a placeholder for future plugin-based postprocessing.
// It returns the document unchanged for now.
func postNormalize(doc *model.Document) *model.Document {
	return doc
}

//
// ────────────────────────────────────────────────────────────────────────
//                            HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// emptyDocument returns a minimal well-formed Document.
func emptyDocument() *model.Document {
	return &model.Document{
		Kind:     model.DocumentKindUnknown,
		Metadata: map[string]string{},
		Sections: nil,
		Content:  "",
		Title:    "",
		Excerpt:  "",
	}
}
