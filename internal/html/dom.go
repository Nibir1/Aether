// internal/html/dom.go
//
// This file contains low-level DOM traversal helpers built on top of
// golang.org/x/net/html. Higher-level extraction functions reuse these
// helpers to locate elements by tag and collect text content.

package html

import (
	"strings"

	xhtml "golang.org/x/net/html"
)

// findElementsByTag finds all element nodes with a given tag name.
func findElementsByTag(n *xhtml.Node, tag string, out *[]*xhtml.Node) {
	if n.Type == xhtml.ElementNode && strings.EqualFold(n.Data, tag) {
		*out = append(*out, n)
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		findElementsByTag(child, tag, out)
	}
}

// textContent returns the concatenated text content of the subtree rooted at n.
func textContent(n *xhtml.Node) string {
	var b strings.Builder
	collectText(n, &b)
	return cleanWhitespace(b.String())
}

// collectText appends text nodes recursively into the builder.
func collectText(n *xhtml.Node, b *strings.Builder) {
	if n.Type == xhtml.TextNode {
		b.WriteString(n.Data)
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		collectText(child, b)
	}
}
