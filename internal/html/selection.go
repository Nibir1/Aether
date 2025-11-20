// internal/html/selection.go
//
// Higher-level selection and extraction helpers built on top of the
// low-level DOM utilities. These functions extract headings and
// paragraphs in a structured way.

package html

import (
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"
)

// Heading represents a heading element (h1–h6).
type Heading struct {
	Level int
	Text  string
}

// Paragraph represents a paragraph of text.
type Paragraph struct {
	Text string
}

// ExtractHeadings extracts all headings (h1–h6) in document order.
func ExtractHeadings(doc *Document) []Heading {
	if doc == nil || doc.Root == nil {
		return nil
	}

	var out []Heading
	var walker func(n *xhtml.Node)
	walker = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && len(n.Data) == 2 && (n.Data[0] == 'h' || n.Data[0] == 'H') {
			level, err := strconv.Atoi(n.Data[1:])
			if err == nil && level >= 1 && level <= 6 {
				text := cleanWhitespace(textContent(n))
				if text != "" {
					out = append(out, Heading{
						Level: level,
						Text:  text,
					})
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walker(child)
		}
	}
	walker(doc.Root)
	return out
}

// ExtractParagraphs extracts <p> elements as Paragraph values.
func ExtractParagraphs(doc *Document) []Paragraph {
	if doc == nil || doc.Root == nil {
		return nil
	}

	var nodes []*xhtml.Node
	findElementsByTag(doc.Root, "p", &nodes)

	out := make([]Paragraph, 0, len(nodes))
	for _, n := range nodes {
		text := cleanWhitespace(textContent(n))
		if text == "" {
			continue
		}
		// Ignore paragraphs that are purely whitespace or boilerplate-like.
		if isBoilerplateParagraph(text) {
			continue
		}
		out = append(out, Paragraph{Text: text})
	}
	return out
}

// isBoilerplateParagraph performs a very shallow heuristic to skip
// paragraphs that are likely navigation or similar noise.
//
// This is intentionally conservative at Stage 4; more advanced scoring
// appears in the ExtractText/Readability stage.
func isBoilerplateParagraph(text string) bool {
	if len(text) < 5 {
		return true
	}
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "© ") || strings.Contains(lower, "cookies") {
		return true
	}
	return false
}
