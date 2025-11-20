// internal/normalize/schema_feed.go
//
// Normalizes Feed data (RSS / Atom) into structured model.Section values.
//
// Feed items become individual sections with Role = feed_item, containing:
//   • heading      – the title of the feed item
//   • text         – best-effort extracted body (content > description > title)
//   • metadata     – link, guid, author, timestamps
//
// The resulting sections are wrapped in a model.Document which is later
// merged into the primary SearchDocument by merge.go.

package normalize

import (
	"strconv"
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

// normalizeFeed converts Feed → model.Document containing N sections.
// merge.go attaches these sections to the core document.
func normalizeFeed(sr *SearchResult) *model.Document {
	if sr == nil || sr.Feed == nil {
		return nil
	}

	f := sr.Feed
	if len(f.Items) == 0 {
		return nil
	}

	sections := make([]model.Section, 0, len(f.Items))

	for _, item := range f.Items {
		heading := strings.TrimSpace(item.Title)
		body := chooseBody(item)

		// Fallbacks
		if heading == "" {
			heading = "(feed item)"
		}
		if body == "" {
			body = heading
		}

		meta := map[string]string{}
		if item.Link != "" {
			meta["link"] = strings.TrimSpace(item.Link)
		}
		if item.Author != "" {
			meta["author"] = strings.TrimSpace(item.Author)
		}
		if item.GUID != "" {
			meta["guid"] = strings.TrimSpace(item.GUID)
		}
		if item.Published != 0 {
			meta["published_unix"] = strconv.FormatInt(item.Published, 10)
		}
		if item.Updated != 0 {
			meta["updated_unix"] = strconv.FormatInt(item.Updated, 10)
		}

		sections = append(sections, model.Section{
			Role:    model.SectionRoleFeedItem,
			Heading: heading,
			Text:    body,
			Meta:    meta,
		})
	}

	// Wrap feed sections in a standalone Document.
	doc := &model.Document{
		Kind:     model.DocumentKindFeed,
		Title:    deriveFeedTitle(sr),
		Excerpt:  "",
		Content:  "",
		Metadata: map[string]string{},
		Sections: sections,
	}

	return doc
}

//
// ────────────────────────────────────────────────────────────────────────
//                             HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// chooseBody picks the best available content field from a feed item.
func chooseBody(item FeedItem) string {
	candidates := []string{
		item.Content,
		item.Description,
		item.Title,
	}

	for _, c := range candidates {
		if strings.TrimSpace(c) != "" {
			return strings.TrimSpace(c)
		}
	}
	return ""
}

// deriveFeedTitle provides a top-level title for the entire feed document.
func deriveFeedTitle(sr *SearchResult) string {
	if sr.PrimaryDocument != nil && sr.PrimaryDocument.Title != "" {
		return strings.TrimSpace(sr.PrimaryDocument.Title)
	}
	return "(feed)"
}
