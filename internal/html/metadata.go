// internal/html/metadata.go
//
// Metadata extraction: <title>, <meta> tags, OpenGraph and Twitter cards.

package html

import (
	"strings"

	xhtml "golang.org/x/net/html"
)

// Meta represents a simple key/value metadata entry.
type Meta struct {
	Name  string
	Value string
}

// ExtractTitle returns the document <title> text, if any.
func ExtractTitle(doc *Document) string {
	if doc == nil || doc.Root == nil {
		return ""
	}

	var titles []*xhtml.Node
	findElementsByTag(doc.Root, "title", &titles)
	if len(titles) == 0 {
		return ""
	}
	return cleanWhitespace(textContent(titles[0]))
}

// ExtractMeta collects <meta> tags into a map.
//
// Keys come from the "name" or "property" attribute. The "content"
// attribute is used as the value. If both name and property are present,
// property takes precedence.
func ExtractMeta(doc *Document) map[string]string {
	if doc == nil || doc.Root == nil {
		return map[string]string{}
	}

	result := make(map[string]string)

	var metaNodes []*xhtml.Node
	findElementsByTag(doc.Root, "meta", &metaNodes)

	for _, node := range metaNodes {
		if node.Type != xhtml.ElementNode {
			continue
		}
		var name, property, content string
		for _, attr := range node.Attr {
			switch strings.ToLower(attr.Key) {
			case "name":
				name = strings.TrimSpace(attr.Val)
			case "property":
				property = strings.TrimSpace(attr.Val)
			case "content":
				content = strings.TrimSpace(attr.Val)
			}
		}
		if content == "" {
			continue
		}
		key := ""
		if property != "" {
			key = property
		} else if name != "" {
			key = name
		}
		if key == "" {
			continue
		}
		result[key] = content
	}
	return result
}
