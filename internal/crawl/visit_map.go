// internal/crawl/visit_map.go
//
// This file implements the visited-set tracking for Aether's crawl subsystem.
// Crawlers must ensure they do not revisit URLs, both for performance and for
// ethical/polite crawling behavior.
//
// A VisitMap stores normalized URLs (canonicalized by the crawler) and ensures
// that each is only visited once.
//
// Design notes:
//   - thread-safe: protected by a mutex
//   - fast: O(1) average-case membership checks
//   - stores only normalized absolute URLs
//   - does not perform normalization itself: crawler is responsible for
//     canonicalizing URLs before inserting them.

package crawl

import "sync"

// VisitMap is a concurrency-safe visited URL registry.
//
// URLs stored here must already be normalized by the crawler subsystem.
// Typically, this includes:
//   - scheme normalization
//   - host lowercasing
//   - path cleaning
//   - removal of URL fragments (#section)
//   - resolution of relative URLs
//
// VisitMap does not perform normalization on its own, by design.
type VisitMap struct {
	mu  sync.Mutex
	set map[string]struct{}
}

// NewVisitMap constructs an empty VisitMap.
func NewVisitMap() *VisitMap {
	return &VisitMap{
		set: make(map[string]struct{}),
	}
}

// MarkVisited records a URL as visited, regardless of prior existence.
//
// It returns true if the URL was newly added, false if it was already present.
func (v *VisitMap) MarkVisited(url string) bool {
	if url == "" {
		return false
	}

	v.mu.Lock()
	_, existed := v.set[url]
	if !existed {
		v.set[url] = struct{}{}
	}
	v.mu.Unlock()

	return !existed
}

// IsVisited reports whether the URL has already been seen.
func (v *VisitMap) IsVisited(url string) bool {
	if url == "" {
		return false
	}

	v.mu.Lock()
	_, ok := v.set[url]
	v.mu.Unlock()

	return ok
}

// Count returns the number of visited URLs so far.
func (v *VisitMap) Count() int {
	v.mu.Lock()
	n := len(v.set)
	v.mu.Unlock()
	return n
}
