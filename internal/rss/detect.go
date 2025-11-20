// internal/rss/detect.go
//
// Lightweight RSS/Atom detection helper used before performing
// a full XML unmarshal. This avoids unnecessary parsing when the
// downloaded content is not actually a feed.
//
// This is separate from Stage 6's MIME/content detection so that
// the RSS subsystem can perform more precise feed identification.

package rss

import (
	"bytes"
	"strings"
)

type FeedType string

const (
	FeedUnknown FeedType = "unknown"
	FeedRSS2    FeedType = "rss2"
	FeedRSS1    FeedType = "rss1" // RDF-based
	FeedAtom    FeedType = "atom"
)

// DetectFeedType performs a lightweight sniff-test on raw bytes to
// determine the likely feed format (RSS 2.0, RSS 1.0, or Atom).
//
// This does *not* perform any XML unmarshalling — it is strictly a
// string-based prefix and tag-name check.
func DetectFeedType(data []byte) FeedType {
	if len(data) == 0 {
		return FeedUnknown
	}

	// Limit inspect region to avoid scanning large feeds.
	inspect := strings.ToLower(string(bytes.TrimSpace(data[:min(len(data), 1024)])))

	// Atom feeds: <feed>
	if strings.Contains(inspect, "<feed") && strings.Contains(inspect, "xmlns=\"http://www.w3.org/2005/atom\"") {
		return FeedAtom
	}

	// RSS 2.0: <rss>
	if strings.Contains(inspect, "<rss") {
		return FeedRSS2
	}

	// RSS 1.0 / RDF: <rdf:RDF> or <rdf>
	if strings.Contains(inspect, "<rdf:rdf") || strings.Contains(inspect, "<rdf") {
		return FeedRSS1
	}

	// Atom variant without explicit namespace
	if strings.Contains(inspect, "<feed") {
		return FeedAtom
	}

	return FeedUnknown
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
