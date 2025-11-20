// internal/extract/scoring.go
//
// This file contains the scoring logic used to identify the main
// article container within an HTML document.

package extract

import (
	"strings"

	xhtml "golang.org/x/net/html"
)

// candidateScore holds scoring information for a DOM node.
type candidateScore struct {
	Node  *xhtml.Node
	Score float64
}

// findBodyNode attempts to locate the <body> element in the document.
func findBodyNode(root *xhtml.Node) *xhtml.Node {
	var body *xhtml.Node
	var walker func(*xhtml.Node)
	walker = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && strings.EqualFold(n.Data, "body") {
			body = n
			return
		}
		for c := n.FirstChild; c != nil && body == nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(root)
	return body
}

// scoreCandidates traverses the DOM and assigns scores to likely
// content-containing nodes.
func scoreCandidates(body *xhtml.Node) []*candidateScore {
	var candidates []*candidateScore
	nodeToScore := make(map[*xhtml.Node]*candidateScore)

	var walker func(*xhtml.Node)
	walker = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode {
			tag := strings.ToLower(n.Data)
			switch tag {
			case "p", "td", "pre", "article", "section", "div", "li":
				text := strings.TrimSpace(nodeText(n))
				if len(text) < 50 {
					break
				}
				base := baseContentScore(tag, text)
				addToNodeScore(nodeToScore, n, base)
				if parent := n.Parent; parent != nil {
					addToNodeScore(nodeToScore, parent, base*0.5)
				}
				if parent := n.Parent; parent != nil && parent.Parent != nil {
					addToNodeScore(nodeToScore, parent.Parent, base*0.25)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(body)

	for _, cs := range nodeToScore {
		// Adjust score by link density: more links => less likely main content.
		density := linkDensity(cs.Node)
		cs.Score = cs.Score * (1.0 - density)
		if cs.Score < 0 {
			cs.Score = 0
		}
		candidates = append(candidates, cs)
	}
	return candidates
}

// baseContentScore calculates an initial content score based on tag and text.
func baseContentScore(tag, text string) float64 {
	score := 0.0

	// Base weight based on tag.
	switch tag {
	case "div", "article", "section":
		score += 5.0
	case "p", "pre", "td", "li":
		score += 3.0
	}

	// Length factor.
	length := float64(len(text))
	score += length / 100.0

	// Comma bonus: paragraphs with commas often indicate substantive text.
	score += float64(strings.Count(text, ","))

	return score
}

func addToNodeScore(m map[*xhtml.Node]*candidateScore, n *xhtml.Node, s float64) {
	if n == nil {
		return
	}
	if existing, ok := m[n]; ok {
		existing.Score += s
	} else {
		m[n] = &candidateScore{
			Node:  n,
			Score: s,
		}
	}
}

// selectTopCandidate returns the node with the highest score.
func selectTopCandidate(candidates []*candidateScore) *xhtml.Node {
	if len(candidates) == 0 {
		return nil
	}
	var best *candidateScore
	for _, c := range candidates {
		if best == nil || c.Score > best.Score {
			best = c
		}
	}
	if best == nil || best.Score <= 0 {
		return nil
	}
	return best.Node
}
