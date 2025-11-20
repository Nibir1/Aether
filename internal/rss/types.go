// internal/rss/types.go
//
// Defines the unified RSS/Atom feed representation used internally.
// Aether normalizes RSS 2.0, RSS 1.0, Atom, and hybrid feeds into a
// consistent structure suitable for LLM and higher-level processing.

package rss

import "time"

// Feed represents a normalized RSS/Atom feed.
type Feed struct {
	Title       string
	Description string
	Link        string
	Updated     time.Time
	Items       []Item
}

// Item represents a single feed entry.
type Item struct {
	Title       string
	Link        string
	Description string
	Content     string
	Author      string
	Published   time.Time
	Updated     time.Time
	GUID        string
}
