// internal/html/links.go
//
// Link extraction helpers. These functions collect anchor tags
// and normalize link text and attributes for higher-level consumers.

package html

import (
	"strings"

	xhtml "golang.org/x/net/html"
)

// Link represents a hyperlink in the document.
type Link struct {
	Href string
	Text string
	Rel  string
}

// ExtractLinks returns all <a> elements as Link values.
func ExtractLinks(doc *Document) []Link {
	if doc == nil || doc.Root == nil {
		return nil
	}

	var nodes []*xhtml.Node
	findElementsByTag(doc.Root, "a", &nodes)

	out := make([]Link, 0, len(nodes))
	for _, n := range nodes {
		var href, rel string
		for _, attr := range n.Attr {
			switch strings.ToLower(attr.Key) {
			case "href":
				href = strings.TrimSpace(attr.Val)
			case "rel":
				rel = strings.TrimSpace(attr.Val)
			}
		}
		text := cleanWhitespace(textContent(n))
		if href == "" && text == "" {
			continue
		}
		out = append(out, Link{
			Href: href,
			Text: text,
			Rel:  rel,
		})
	}

	return out
}
