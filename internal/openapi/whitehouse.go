// internal/openapi/whitehouse.go
//
// White House posts integration using the public WordPress JSON API:
//   https://www.whitehouse.gov/wp-json/wp/v2/posts?per_page=N
//
// This provides recent posts, which often include press releases,
// blog posts, and other official communications. No authentication
// is required.

package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

// WhiteHousePost represents a normalized White House post.
type WhiteHousePost struct {
	ID      int64
	Title   string
	URL     string
	Date    time.Time
	Excerpt string
}

// whiteHousePostRaw matches the subset of the WP JSON schema we use.
type whiteHousePostRaw struct {
	ID    int64  `json:"id"`
	Link  string `json:"link"`
	Date  string `json:"date"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Excerpt struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
}

// WhiteHouseRecentPosts retrieves the latest N posts from
// whitehouse.gov using the WP JSON API.
func (c *Client) WhiteHouseRecentPosts(ctx context.Context, limit int) ([]WhiteHousePost, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	endpoint := fmt.Sprintf("https://www.whitehouse.gov/wp-json/wp/v2/posts?per_page=%d", limit)

	body, _, err := c.getJSON(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var raw []whiteHousePostRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	out := make([]WhiteHousePost, 0, len(raw))
	for _, r := range raw {
		t, _ := time.Parse(time.RFC3339, r.Date)
		excerptText := stripHTML(r.Excerpt.Rendered)
		titleText := stripHTML(r.Title.Rendered)

		out = append(out, WhiteHousePost{
			ID:      r.ID,
			Title:   titleText,
			URL:     r.Link,
			Date:    t,
			Excerpt: excerptText,
		})
	}

	return out, nil
}

// stripHTML removes basic HTML tags from a string and unescapes entities.
var tagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)
	return s
}
