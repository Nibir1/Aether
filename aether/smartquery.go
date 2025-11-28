// aether/smartquery.go
//
// This file exposes Aether's SmartQuery classification and routing API.
// It does not perform network calls; instead it analyzes a query string,
// infers intent, and returns a routing plan that higher-level functions
// (such as Search) can execute.
// SmartQuery API — classification + routing plan generator.
// Pure logic: no I/O, no network calls, deterministic and unit-testable.
// This keeps SmartQuery deterministic, fast, and easy to unit-test.

package aether

import (
	"strings"

	ismart "github.com/Nibir1/Aether/internal/smartquery"
)

// QueryIntent mirrors the internal smartquery.Intent type.
type QueryIntent string

const (
	QueryIntentUnknown       QueryIntent = "unknown"
	QueryIntentGeneralSearch QueryIntent = "general_search"
	QueryIntentLookup        QueryIntent = "lookup"
	QueryIntentNews          QueryIntent = "news"
	QueryIntentDocs          QueryIntent = "docs"
	QueryIntentCodeHelp      QueryIntent = "code_help"
	QueryIntentRSS           QueryIntent = "rss"
	QueryIntentHackerNews    QueryIntent = "hackernews"
	QueryIntentGitHub        QueryIntent = "github"
)

// SmartQueryPlan is the public routing plan returned by Aether.
type SmartQueryPlan struct {
	Query           string
	Intent          QueryIntent
	IsQuestion      bool
	HasURL          bool
	PrimarySources  []string
	FallbackSources []string
	UseLookup       bool
	UseSearchIndex  bool
	UseOpenAPIs     bool
	UseFeeds        bool
	UsePlugins      bool
}

// SmartQuery analyzes a natural-language query and returns a routing plan.
// Pure function. No network calls.
func (c *Client) SmartQuery(query string) *SmartQueryPlan {
	// Nil-client still produces a valid deterministic plan.
	if c == nil {
		trimmed := strings.TrimSpace(query)
		return &SmartQueryPlan{
			Query:  trimmed,
			Intent: QueryIntentUnknown,
		}
	}

	trimmed := strings.TrimSpace(query)

	// Internal classification (struct)
	internalClass := ismart.Classify(trimmed)

	// Routing decision (struct)
	route := ismart.BuildRoute(internalClass)

	// Defensive deep-copy of slices (struct slices, so can still be nil)
	primary := append([]string(nil), route.PrimarySources...)
	fallback := append([]string(nil), route.FallbackSources...)

	return &SmartQueryPlan{
		Query:           trimmed,
		Intent:          mapIntent(internalClass.Intent),
		IsQuestion:      internalClass.IsQuestion,
		HasURL:          internalClass.HasURL,
		PrimarySources:  primary,
		FallbackSources: fallback,
		UseLookup:       route.UseLookup,
		UseSearchIndex:  route.UseSearchIndex,
		UseOpenAPIs:     route.UseOpenAPIs,
		UseFeeds:        route.UseFeeds,
		UsePlugins:      route.UsePlugins,
	}
}

// mapIntent converts internal Intent (struct value) to public QueryIntent.
func mapIntent(i ismart.Intent) QueryIntent {
	switch i {
	case ismart.IntentLookup:
		return QueryIntentLookup
	case ismart.IntentNews:
		return QueryIntentNews
	case ismart.IntentDocs:
		return QueryIntentDocs
	case ismart.IntentCodeHelp:
		return QueryIntentCodeHelp
	case ismart.IntentRSS:
		return QueryIntentRSS
	case ismart.IntentHackerNews:
		return QueryIntentHackerNews
	case ismart.IntentGitHub:
		return QueryIntentGitHub
	case ismart.IntentGeneralSearch:
		return QueryIntentGeneralSearch
	default:
		return QueryIntentUnknown
	}
}
