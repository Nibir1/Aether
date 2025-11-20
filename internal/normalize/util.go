// internal/normalize/util.go
//
// Shared helpers for the normalization subsystem.
//
// These functions support:
//
//   • consistent whitespace collapsing
//   • safe metadata cloning
//   • excerpt trimming
//   • safe string operations
//   • section de-duplication (for feeds, entities, etc.)
//   • URL cleanup and fallback helpers
//
// The goal is to centralize common logic so schema_* files remain focused
// on their domain-specific transformations.

package normalize

import (
	"regexp"
	"strings"

	"github.com/Nibir1/Aether/internal/model"
)

//
// ────────────────────────────────────────────────────────────────────────
//                            WHITESPACE HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// collapseWhitespace reduces all internal whitespace sequences to a single
// space and trims surrounding whitespace. Useful for excerpts.
func collapseWhitespace(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Replace any sequence of whitespace with a single space.
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

// safeTrim returns trimmed s, preserving empty result safely.
func safeTrim(s string) string {
	return strings.TrimSpace(s)
}

//
// ────────────────────────────────────────────────────────────────────────
//                             METADATA HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// safeMetadataCopy clones m into a new map. Nil-safe.
func safeMetadataCopy(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k != "" {
			out[k] = v
		}
	}
	return out
}

// mergeMetadata merges b with a, preserving keys in base.
//
// Base keys are never overwritten. Overlay supplies missing keys.
func mergeMetadata(base, overlay map[string]string) map[string]string {
	if base == nil && overlay == nil {
		return nil
	}
	if base == nil {
		return safeMetadataCopy(overlay)
	}
	if overlay == nil {
		return base
	}

	for k, v := range overlay {
		if _, exists := base[k]; !exists {
			base[k] = v
		}
	}
	return base
}

//
// ────────────────────────────────────────────────────────────────────────
//                           SECTION HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// dedupeSections removes exact-duplicate sections based on Role, Heading,
// and Text. Metadata is *not* considered for dedupe criteria.
func dedupeSections(sections []model.Section) []model.Section {
	seen := map[string]struct{}{}
	out := make([]model.Section, 0, len(sections))

	for _, s := range sections {
		// Cast s.Role (model.SectionRole) to string for concatenation.
		key := string(s.Role) + "|" + s.Heading + "|" + s.Text
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}

	return out
}

//
// ────────────────────────────────────────────────────────────────────────
//                              URL HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// deriveTitleFromURL provides a fallback title based on the final segment
// of the URL path. Used only as a last resort.
func deriveTitleFromURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return "(untitled)"
	}

	parts := strings.Split(url, "/")
	last := strings.TrimSpace(parts[len(parts)-1])
	if last == "" {
		return url
	}
	return last
}

//
// ────────────────────────────────────────────────────────────────────────
//                        EXCERPT / CONTENT HELPERS
// ────────────────────────────────────────────────────────────────────────
//

// excerptFromContent returns up to limit characters from the content.
// If limit <= 0, defaults to 240 chars.
func excerptFromContent(content string, limit int) string {
	content = collapseWhitespace(content)
	if content == "" {
		return ""
	}
	if limit <= 0 {
		limit = 240
	}
	if len(content) <= limit {
		return content
	}
	return content[:limit] + "…"
}

//
// ────────────────────────────────────────────────────────────────────────
//                       DOCUMENT / SECTION BUILDERS
// ────────────────────────────────────────────────────────────────────────
//

// newSection creates a Section with role, heading, text, and metadata.
func newSection(role model.SectionRole, heading, text string, meta map[string]string) model.Section {
	return model.Section{
		Role:    role,
		Heading: safeTrim(heading),
		Text:    safeTrim(text),
		Meta:    safeMetadataCopy(meta),
	}
}
