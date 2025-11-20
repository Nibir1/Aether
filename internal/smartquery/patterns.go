// internal/smartquery/patterns.go
//
// This file defines keyword and heuristic patterns used by the SmartQuery
// classifier. The goal is to recognize user intent (news, docs, lookup,
// code help, RSS, Hacker News, etc.) in a lightweight, deterministic way.

package smartquery

import "strings"

// simple keyword sets for different intents.
// These are intentionally small and explainable. They can be extended
// over time without changing the public API.

var newsKeywords = []string{
	"news", "headline", "breaking", "latest", "today", "update",
}

var docsKeywords = []string{
	"documentation", "docs", "reference", "api reference", "manual", "guide", "how to use",
}

var definitionKeywords = []string{
	"what is", "who is", "define ", "definition of", "meaning of",
}

var codeKeywords = []string{
	"error:", "stack trace", "exception", "how to fix", "how do i", "segmentation fault",
	"nullpointer", "undefined reference", "compile error",
}

var rssKeywords = []string{
	"rss feed", "atom feed", "subscribe", "feed url",
}

var hnKeywords = []string{
	"hacker news", "hn top", "hn best", "hn new",
}

var lookupKeywords = []string{
	"wiki", "wikipedia", "quick fact", "lookup", "short answer",
}

var githubKeywords = []string{
	"github repo", "github repository", "awesome list",
}

// containsAny reports whether s contains any of the given substrings (case-insensitive).
func containsAny(s string, list []string) bool {
	lower := strings.ToLower(s)
	for _, k := range list {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// looksLikeURL tries to determine if the query is (mostly) a URL.
func looksLikeURL(q string) bool {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return false
	}
	if strings.HasPrefix(q, "http://") || strings.HasPrefix(q, "https://") {
		return true
	}
	if strings.Contains(q, ".") && !strings.Contains(q, " ") {
		// simple hostname or bare domain
		return true
	}
	return false
}

// looksLikeQuestion checks if the query is phrased as a question.
func looksLikeQuestion(q string) bool {
	lower := strings.ToLower(strings.TrimSpace(q))
	if strings.HasSuffix(lower, "?") {
		return true
	}
	if strings.HasPrefix(lower, "what ") ||
		strings.HasPrefix(lower, "who ") ||
		strings.HasPrefix(lower, "how ") ||
		strings.HasPrefix(lower, "why ") ||
		strings.HasPrefix(lower, "when ") {
		return true
	}
	return false
}
