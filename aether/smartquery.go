// aether/smartquery.go
//
// This file exposes Aether's SmartQuery classification and routing API.
// It does not perform network calls; instead it analyzes a query string,
// infers intent, and returns a routing plan that higher-level functions
// (such as Search) can execute.
//
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
//
// It describes how Aether *would* answer the query, without actually
// performing the network calls.
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
//
// This is a pure function of the query string: it does not hit the network,
// does not fetch content, and does not depend on remote state. It is safe
// to call frequently and suitable for unit testing.
func (c *Client) SmartQuery(query string) *SmartQueryPlan {
	trimmed := strings.TrimSpace(query)
	internalClass := ismart.Classify(trimmed)
	route := ismart.BuildRoute(internalClass)

	return &SmartQueryPlan{
		Query:           trimmed,
		Intent:          mapIntent(internalClass.Intent),
		IsQuestion:      internalClass.IsQuestion,
		HasURL:          internalClass.HasURL,
		PrimarySources:  append([]string(nil), route.PrimarySources...),
		FallbackSources: append([]string(nil), route.FallbackSources...),
		UseLookup:       route.UseLookup,
		UseSearchIndex:  route.UseSearchIndex,
		UseOpenAPIs:     route.UseOpenAPIs,
		UseFeeds:        route.UseFeeds,
		UsePlugins:      route.UsePlugins,
	}
}

// mapIntent converts internal Intent to public QueryIntent.
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
	case ismart.IntentUnknown:
		fallthrough
	default:
		return QueryIntentUnknown
	}
}
