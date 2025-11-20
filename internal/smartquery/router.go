// internal/smartquery/router.go
//
// This file implements the SmartQuery routing logic. Given a
// Classification, it produces a Route describing which logical
// subsystems Aether should consult: Lookup, Open APIs, RSS, HN, etc.
//
// Stage 7 does not yet *execute* these routes; it only plans them.
// Later stages (Search/OpenAPIs) will consume this routing information.

package smartquery

// Route describes which logical Aether sources should be consulted
// to answer a query. The names here are abstract and correspond to
// internal modules or plugins (e.g. "lookup", "wikipedia", "rss").
type Route struct {
	Intent          Intent
	PrimarySources  []string
	FallbackSources []string
	UseLookup       bool
	UseSearchIndex  bool
	UseOpenAPIs     bool
	UseFeeds        bool
	UsePlugins      bool
}

// BuildRoute constructs a routing plan based on the classification.
func BuildRoute(c Classification) Route {
	r := Route{
		Intent:          c.Intent,
		PrimarySources:  nil,
		FallbackSources: nil,
		UseLookup:       false,
		UseSearchIndex:  false,
		UseOpenAPIs:     false,
		UseFeeds:        false,
		UsePlugins:      false,
	}

	switch c.Intent {
	case IntentLookup:
		r.UseLookup = true
		r.PrimarySources = []string{"lookup", "wikipedia", "wikidata"}
		r.FallbackSources = []string{"openapi:github", "openapi:gov"}
	case IntentNews:
		r.UseFeeds = true
		r.UseOpenAPIs = true
		r.PrimarySources = []string{"rss", "openapi:hackernews"}
		r.FallbackSources = []string{"openapi:news", "openapi:gov"}
	case IntentDocs:
		r.UseOpenAPIs = true
		r.UseSearchIndex = true
		r.PrimarySources = []string{"docs:index", "openapi:github"}
		r.FallbackSources = []string{"search:web"}
	case IntentCodeHelp:
		r.UseSearchIndex = true
		r.UseOpenAPIs = true
		r.PrimarySources = []string{"stacktrace", "openapi:github"}
		r.FallbackSources = []string{"search:web"}
	case IntentHackerNews:
		r.UseOpenAPIs = true
		r.PrimarySources = []string{"openapi:hackernews"}
		r.FallbackSources = []string{"rss", "search:web"}
	case IntentGitHub:
		r.UseOpenAPIs = true
		r.PrimarySources = []string{"openapi:github"}
		r.FallbackSources = []string{"search:web"}
	case IntentGeneralSearch:
		r.UseSearchIndex = true
		r.UseOpenAPIs = true
		r.PrimarySources = []string{"search:web"}
		r.FallbackSources = []string{"lookup", "openapi:wikipedia"}
	case IntentUnknown:
		fallthrough
	default:
		// Unknown intents fall back to general search and lightweight lookup.
		r.UseSearchIndex = true
		r.UseLookup = true
		r.PrimarySources = []string{"search:web"}
		r.FallbackSources = []string{"lookup"}
	}

	// Plugins are allowed for all intents; plugins decide if they match.
	r.UsePlugins = true

	return r
}
