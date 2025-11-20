// internal/openapi/press.go
//
// Press-release integrations using free and public sources.
// This includes:
//   - GDELT GKG 2.0 "press release" filtered queries
//   - Direct government RSS feeds (e.g., EU, UK, CA)
//
// No API key required.

package openapi

import (
	"context"
	"encoding/xml"
	"strings"
	"time"
)

// PressRelease represents a normalized government press item.
type PressRelease struct {
	Title   string
	URL     string
	Source  string
	Snippet string
	Date    time.Time
}

// Basic RSS entry used by many gov feeds.
type govRSS struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

// GovernmentPressFeeds is a curated set of RSS feeds for
// legal, public, API-free press releases.
var GovernmentPressFeeds = []string{
	"https://www.whitehouse.gov/feed/",
	"https://www.gov.uk/government/announcements.atom",
	"https://news.gc.ca/web/fd-en.do?format=rss",          // Canada
	"https://ec.europa.eu/commission/presscorner/home/en", // EU
}

// GovernmentPress retrieves recent press releases from major agencies.
func (c *Client) GovernmentPress(ctx context.Context, limit int) ([]PressRelease, error) {
	if limit <= 0 {
		limit = 10
	}

	out := []PressRelease{}

	for _, feedURL := range GovernmentPressFeeds {
		body, _, err := c.getText(ctx, feedURL)
		if err != nil {
			continue
		}

		var f govRSS
		if err := xml.Unmarshal(body, &f); err != nil {
			continue
		}

		for _, it := range f.Channel.Items {
			if len(out) >= limit {
				return out, nil
			}

			t, _ := time.Parse(time.RFC1123Z, it.PubDate)

			out = append(out, PressRelease{
				Title:   strings.TrimSpace(it.Title),
				URL:     it.Link,
				Source:  feedURL,
				Snippet: strings.TrimSpace(stripHTML(it.Description)),
				Date:    t,
			})
		}
	}

	return out, nil
}
