// aether/rss.go
//
// Public RSS/Atom feed fetching and parsing interface for Aether.
// This integrates Aether’s robots.txt-compliant HTTP fetcher with the
// internal RSS/Atom parsers and feed-type sniffing logic.
//
// Stage 8:
//  - Adds full RSS/Atom support
//  - Automatically detects feed type before parsing
//  - Normalizes all feed fields into stable LLM-friendly structures

package aether

import (
	"context"
	"errors"

	irss "github.com/Nibir1/Aether/internal/rss"
)

// FeedItem is the public representation of a feed entry.
type FeedItem struct {
	Title       string
	Link        string
	Description string
	Content     string
	Author      string
	Published   int64
	Updated     int64
	GUID        string
}

// Feed is the public normalized RSS/Atom feed.
type Feed struct {
	Title       string
	Description string
	Link        string
	Updated     int64
	Items       []FeedItem
}

// ParseRSS parses raw RSS/Atom XML bytes into a public Feed.
//
// This method does NOT fetch or check robots.txt; it only parses.
// Use FetchRSS() to fetch and parse in one call.
func (c *Client) ParseRSS(xmlBytes []byte) (*Feed, error) {

	// Step 1 — fast pre-check using DetectFeedType
	ft := irss.DetectFeedType(xmlBytes)
	if ft == irss.FeedUnknown {
		return nil, errors.New("aether: content does not appear to be a valid RSS/Atom feed")
	}

	// Step 2 — full parse of RSS/Atom variants
	internalFeed, err := irss.Parse(xmlBytes)
	if err != nil {
		return nil, err
	}

	// Step 3 — cleanup / normalization
	internalFeed.Clean()

	// Step 4 — convert to public type
	out := &Feed{
		Title:       internalFeed.Title,
		Description: internalFeed.Description,
		Link:        internalFeed.Link,
		Updated:     internalFeed.Updated.Unix(),
	}

	for _, it := range internalFeed.Items {
		out.Items = append(out.Items, FeedItem{
			Title:       it.Title,
			Link:        it.Link,
			Description: it.Description,
			Content:     it.Content,
			Author:      it.Author,
			Published:   it.Published.Unix(),
			Updated:     it.Updated.Unix(),
			GUID:        it.GUID,
		})
	}

	return out, nil
}

// FetchRSS fetches and parses an RSS/Atom feed, respecting robots.txt.
//
// This is the safe and preferred method for real-world usage.
//
// Example:
//
//	feed, err := client.FetchRSS(ctx, "https://example.com/feed.rss")
func (c *Client) FetchRSS(ctx context.Context, url string) (*Feed, error) {
	resp, err := c.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	return c.ParseRSS(resp.Body)
}
