// internal/normalize/schema_article.go
//
// Normalizes Article extraction (Readability-style content) into
// structured model.Section values, which are later merged into the
// canonical Document by the merge.go layer.
//
// An Article contains cleaned HTML → text extraction:
//   Title
//   Content
//   Meta (map[string]string)
//
// Normalization rules:
//   • Produces exactly one main body section.
//   • The SearchResult.PrimaryDocument establishes the root title,
//     but Article content supersedes it as richer content.
//   • Article.Meta is preserved as section-level metadata.

package normalize

import (
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// normalizeArticle converts sr.Article into a single Section and
// wraps it in a model.Document. merge.go will fold it into the
// parent Document created by normalizeSearchDocument.
func normalizeArticle(sr *SearchResult) *model.Document {
	if sr == nil || sr.Article == nil {
		return nil
	}

	art := sr.Article

	title := strings.TrimSpace(art.Title)
	content := strings.TrimSpace(art.Content)

	// If no content, nothing to normalize
	if content == "" {
		return nil
	}

	// Derive fallback title if needed
	if title == "" {
		if sr.PrimaryDocument != nil && sr.PrimaryDocument.Title != "" {
			title = sr.PrimaryDocument.Title
		} else {
			title = "(article)"
		}
	}

	section := model.Section{
		Role:    model.SectionRoleBody,
		Heading: title,
		Text:    content,
		Meta:    copyMetadata(art.Meta),
	}

	doc := &model.Document{
		Kind:     model.DocumentKindArticle,
		Title:    title,
		Excerpt:  deriveExcerpt(content),
		Content:  content,
		Metadata: map[string]string{}, // no root metadata; section metadata only
		Sections: []model.Section{section},
	}

	return doc
}

//
// ────────────────────────────────────────────────────────────────────────
//                               HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// copyMetadata safely clones a map to avoid mutation.
func copyMetadata(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}
