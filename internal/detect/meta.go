// internal/detect/meta.go
//
// Metadata extraction for <title>, <meta>, OpenGraph, Twitter, canonical URL.
// This is a lightweight version used BEFORE readability extraction.

package detect

import (
	ihtml "github.com/Nibir1/Aether/internal/html"
)

// ExtractBasicMeta returns a map of metadata extracted from a Document.
func ExtractBasicMeta(doc *ihtml.Document) map[string]string {
	meta := ihtml.ExtractMeta(doc)
	title := ihtml.ExtractTitle(doc)

	if title != "" {
		meta["title"] = title
	}

	// Promote common OG/Twitter keys
	if m := meta["og:title"]; m != "" {
		meta["title"] = m
	}
	if m := meta["twitter:title"]; m != "" {
		meta["title"] = m
	}
	if m := meta["og:description"]; m != "" {
		meta["description"] = m
	}
	if m := meta["twitter:description"]; m != "" {
		meta["description"] = m
	}

	if m := meta["canonical"]; m != "" {
		meta["canonical_url"] = m
	}
	if m := meta["og:url"]; m != "" {
		meta["canonical_url"] = m
	}

	return meta
}
