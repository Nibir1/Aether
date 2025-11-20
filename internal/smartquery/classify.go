// internal/smartquery/classify.go
//
// This file implements the core SmartQuery classification logic.
// Given a natural-language query string, it infers user intent such as:
//
//   - General search
//   - Quick factual lookup
//   - News / RSS
//   - Technical docs
//   - Code / error debugging
//   - Hacker News
//
// The classifier is intentionally simple and deterministic. Higher-level
// routing is implemented separately in router.go.

package smartquery

import "strings"

// Intent represents the high-level intent of a query.
type Intent string

const (
	IntentUnknown       Intent = "unknown"
	IntentGeneralSearch Intent = "general_search"
	IntentLookup        Intent = "lookup"
	IntentNews          Intent = "news"
	IntentDocs          Intent = "docs"
	IntentCodeHelp      Intent = "code_help"
	IntentRSS           Intent = "rss"
	IntentHackerNews    Intent = "hackernews"
	IntentGitHub        Intent = "github"
)

// Classification is the internal result of query analysis.
type Classification struct {
	Raw        string   // original user query
	Intent     Intent   // inferred high-level intent
	IsQuestion bool     // true if the query looks like a question
	HasURL     bool     // true if the query appears to contain a URL
	Keywords   []string // optional list of detected intent keywords
}

// Classify analyzes a natural-language query and returns an internal
// classification. It uses simple heuristics and keyword patterns and
// does not perform any network calls.
func Classify(query string) Classification {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)

	c := Classification{
		Raw:        q,
		Intent:     IntentUnknown,
		IsQuestion: looksLikeQuestion(q),
		HasURL:     looksLikeURL(q),
		Keywords:   nil,
	}

	// Pure URL queries often mean "analyze this page" or "extract from this URL".
	if c.HasURL && len(strings.Fields(q)) == 1 {
		// Intent will be interpreted as general_search or article extraction later;
		// keep IntentUnknown but HasURL=true as an explicit signal.
		return c
	}

	// Definitions / fact lookup
	if containsAny(lower, definitionKeywords) || containsAny(lower, lookupKeywords) {
		c.Intent = IntentLookup
		return c
	}

	// News / RSS
	if containsAny(lower, newsKeywords) || containsAny(lower, rssKeywords) {
		c.Intent = IntentNews
		return c
	}

	// Hacker News specific
	if containsAny(lower, hnKeywords) {
		c.Intent = IntentHackerNews
		return c
	}

	// Documentation / API reference
	if containsAny(lower, docsKeywords) {
		c.Intent = IntentDocs
		return c
	}

	// Code / error debugging
	if containsAny(lower, codeKeywords) {
		c.Intent = IntentCodeHelp
		return c
	}

	// GitHub-related intents
	if containsAny(lower, githubKeywords) {
		c.Intent = IntentGitHub
		return c
	}

	// Fallback: question-like queries become general search/QA.
	if c.IsQuestion {
		c.Intent = IntentGeneralSearch
		return c
	}

	// Short single-word or two-word queries often map to lookup.
	if len(strings.Fields(q)) <= 2 && !c.HasURL {
		c.Intent = IntentLookup
		return c
	}

	// Default: general search.
	c.Intent = IntentGeneralSearch
	return c
}
