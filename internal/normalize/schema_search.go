// internal/normalize/schema_search.go
//
// Normalization logic for SearchDocument → model.Document.
// This is the primary schema used in Aether's normalization pipeline.
//
// The goal of this file is to convert a SearchDocument—Aether’s unified
// representation of raw fetched content—into the canonical internal/model.Document
// used for JSON and TOON output.
//
// Rules:
//   • Title, excerpt, and content fallbacks are applied when missing.
//   • Metadata is copied and sanitized.
//   • DocumentKind is mapped using internal conventions.
//   • PrimaryDocument becomes the structural foundation onto which
//     article/feed/entity layers may later merge.

package normalize

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// normalizeSearchDocument converts the SearchResult’s PrimaryDocument
// into a canonical model.Document.
//
// Later pipeline stages (merge.go) will attach article sections,
// feed items, and entity sections onto the result of this function.
func normalizeSearchDocument(sr *SearchResult) *model.Document {
	doc := sr.PrimaryDocument
	if doc == nil {
		return emptyDocument()
	}

	kind := mapSearchKind(doc.Kind)

	// Metadata copy
	meta := map[string]string{}
	for k, v := range doc.Metadata {
		meta[k] = strings.TrimSpace(v)
	}

	// Apply "best effort" fallbacks
	title := strings.TrimSpace(doc.Title)
	content := strings.TrimSpace(doc.Content)
	excerpt := strings.TrimSpace(doc.Excerpt)

	// If excerpt missing → derive from content
	if excerpt == "" && content != "" {
		excerpt = deriveExcerpt(content)
	}

	// If title missing → fallback to excerpt or URL
	if title == "" {
		switch {
		case excerpt != "":
			title = excerpt
		case doc.URL != "":
			title = doc.URL
		default:
			title = "(untitled)"
		}
	}

	// Trim everything
	title = strings.TrimSpace(title)
	excerpt = strings.TrimSpace(excerpt)
	content = strings.TrimSpace(content)

	// Construct the normalized document
	normalized := &model.Document{
		SourceURL: doc.URL,
		Kind:      kind,
		Title:     title,
		Excerpt:   excerpt,
		Content:   content,
		Metadata:  meta,
		Sections:  nil, // other schemas attach sections later
	}

	return normalized
}

//
// ────────────────────────────────────────────────────────────────────────
//                            KIND MAPPING
// ────────────────────────────────────────────────────────────────────────
//

func mapSearchKind(kind string) model.DocumentKind {
	switch strings.ToLower(kind) {
	case "article":
		return model.DocumentKindArticle
	case "feed":
		return model.DocumentKindFeed
	case "html":
		return model.DocumentKindHTML
	case "json":
		return model.DocumentKindJSON
	case "text":
		return model.DocumentKindText
	case "binary":
		return model.DocumentKindBinary
	default:
		return model.DocumentKindUnknown
	}
}

//
// ────────────────────────────────────────────────────────────────────────
//                             EXCERPT LOGIC
// ────────────────────────────────────────────────────────────────────────
//

// deriveExcerpt extracts a snippet from the content for preview purposes.
// This function is intentionally simple—Aether aims for clarity over
// linguistic complexity.
func deriveExcerpt(content string) string {
	clean := strings.TrimSpace(content)
	if clean == "" {
		return ""
	}

	// Use first 240 characters (soft heuristic)
	if len(clean) <= 240 {
		return clean
	}
	return clean[:240] + "…"
}
