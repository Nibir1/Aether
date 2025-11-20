// internal/extract/cleaner.go
//
// Cleaner and helper routines to support article extraction.

package extract

import (
	"strings"
	"unicode"

	xhtml "golang.org/x/net/html"
)

// cleanNodeTree removes or skips elements that are unlikely to be part
// of the main content, such as <script>, <style>, <nav>, <aside>, etc.
func cleanNodeTree(root *xhtml.Node) {
	if root == nil {
		return
	}

	var walker func(*xhtml.Node)
	walker = func(n *xhtml.Node) {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			if shouldDropNode(c) {
				// Remove node from tree.
				if c.PrevSibling != nil {
					c.PrevSibling.NextSibling = c.NextSibling
				} else {
					n.FirstChild = c.NextSibling
				}
				if c.NextSibling != nil {
					c.NextSibling.PrevSibling = c.PrevSibling
				} else {
					n.LastChild = c.PrevSibling
				}
			} else {
				walker(c)
			}
			c = next
		}
	}
	walker(root)
}

// shouldDropNode decides whether to remove a node as pure boilerplate.
func shouldDropNode(n *xhtml.Node) bool {
	if n.Type != xhtml.ElementNode {
		return false
	}
	tag := strings.ToLower(n.Data)
	switch tag {
	case "script", "style", "noscript", "iframe", "footer", "nav", "aside", "header", "form":
		return true
	}
	// Heuristic based on class/id hints.
	classID := strings.ToLower(nodeClassAndID(n))
	if classID == "" {
		return false
	}
	if strings.Contains(classID, "comment") ||
		strings.Contains(classID, "footer") ||
		strings.Contains(classID, "sidebar") ||
		strings.Contains(classID, "nav") ||
		strings.Contains(classID, "menu") ||
		strings.Contains(classID, "advert") ||
		strings.Contains(classID, "ad-") {
		return true
	}
	return false
}

// nodeClassAndID returns the concatenation of class and id attributes.
func nodeClassAndID(n *xhtml.Node) string {
	var parts []string
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "class", "id":
			parts = append(parts, a.Val)
		}
	}
	return strings.Join(parts, " ")
}

// nodeText extracts plain text for a node, collapsing whitespace.
func nodeText(n *xhtml.Node) string {
	var b strings.Builder
	collectText(n, &b)
	return collapseWhitespace(b.String())
}

// collectText gathers text nodes recursively, skipping script/style.
func collectText(n *xhtml.Node, b *strings.Builder) {
	if n.Type == xhtml.TextNode {
		b.WriteString(n.Data)
	}
	if n.Type == xhtml.ElementNode {
		tag := strings.ToLower(n.Data)
		if tag == "script" || tag == "style" {
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, b)
	}
}

// collapseWhitespace reduces whitespace sequences to single spaces.
func collapseWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// linkDensity computes the ratio of text inside links vs total text.
func linkDensity(n *xhtml.Node) float64 {
	totalText := len(nodeText(n))
	if totalText == 0 {
		return 0
	}
	linkText := 0

	var walker func(*xhtml.Node)
	walker = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode && strings.EqualFold(node.Data, "a") {
			linkText += len(nodeText(node))
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(n)

	return float64(linkText) / float64(totalText)
}

// buildContentNode returns a shallow clone of the top candidate node
// to serve as the root of the extracted article fragment.
//
// Future stages may expand this to include carefully selected siblings.
func buildContentNode(top *xhtml.Node) *xhtml.Node {
	if top == nil {
		return nil
	}
	clone := shallowClone(top)
	clone.FirstChild, clone.LastChild = nil, nil

	for c := top.FirstChild; c != nil; c = c.NextSibling {
		if isContentChild(c) {
			clone.AppendChild(deepClone(c))
		}
	}
	return clone
}

// isContentChild decides whether a child node is likely to be part of
// the main content block.
func isContentChild(n *xhtml.Node) bool {
	if n.Type == xhtml.ElementNode {
		tag := strings.ToLower(n.Data)
		switch tag {
		case "p", "div", "article", "section", "ul", "ol", "li", "img", "figure", "h1", "h2", "h3", "h4", "h5", "h6":
			return true
		}
	}
	if n.Type == xhtml.TextNode {
		return strings.TrimSpace(n.Data) != ""
	}
	return false
}

// shallowClone clones a node without its children.
func shallowClone(n *xhtml.Node) *xhtml.Node {
	c := &xhtml.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     append([]xhtml.Attribute(nil), n.Attr...),
	}
	return c
}

// deepClone clones a node and its entire subtree.
func deepClone(n *xhtml.Node) *xhtml.Node {
	root := shallowClone(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childClone := deepClone(c)
		root.AppendChild(childClone)
	}
	return root
}

// makeExcerpt produces a short excerpt from the article text.
func makeExcerpt(text string) string {
	if text == "" {
		return ""
	}
	if len(text) <= 240 {
		return text
	}
	// Cut at a word boundary near 240 chars.
	runes := []rune(text)
	if len(runes) <= 240 {
		return text
	}
	runes = runes[:240]
	s := string(runes)
	lastSpace := strings.LastIndex(s, " ")
	if lastSpace > 80 {
		s = s[:lastSpace]
	}
	return strings.TrimSpace(s) + "…"
}
