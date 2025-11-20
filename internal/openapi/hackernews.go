// internal/openapi/hackernews.go
//
// Hacker News integration using the public Firebase API:
//   https://hacker-news.firebaseio.com/v0/topstories.json
//   https://hacker-news.firebaseio.com/v0/item/{id}.json
//
// This module exposes helper methods for retrieving top stories,
// which are useful for news-oriented queries.

package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// HNStory represents a normalized Hacker News story.
type HNStory struct {
	ID           int64
	Title        string
	URL          string
	Author       string
	Score        int
	Time         time.Time
	CommentCount int
}

// hnItemResponse matches the HN item JSON structure (subset).
type hnItemResponse struct {
	ID    int64   `json:"id"`
	Title string  `json:"title"`
	URL   string  `json:"url"`
	By    string  `json:"by"`
	Score int     `json:"score"`
	Time  int64   `json:"time"`
	Kids  []int64 `json:"kids"`
	Type  string  `json:"type"`
}

// HackerNewsTopStories retrieves the top N Hacker News stories.
// limit is clamped to a sensible maximum to avoid excessive calls.
func (c *Client) HackerNewsTopStories(ctx context.Context, limit int) ([]HNStory, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	body, _, err := c.getJSON(ctx, "https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, err
	}

	if len(ids) < limit {
		limit = len(ids)
	}

	out := make([]HNStory, 0, limit)
	for i := 0; i < limit; i++ {
		id := ids[i]
		item, err := c.hnFetchItem(ctx, id)
		if err != nil {
			// Skip failed items and continue.
			continue
		}
		if item == nil {
			continue
		}
		out = append(out, *item)
	}
	return out, nil
}

// hnFetchItem fetches and normalizes a single HN item.
func (c *Client) hnFetchItem(ctx context.Context, id int64) (*HNStory, error) {
	endpoint := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	body, _, err := c.getJSON(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp hnItemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	// Only keep "story" items.
	if resp.Type != "story" {
		return nil, nil
	}

	story := &HNStory{
		ID:     resp.ID,
		Title:  resp.Title,
		URL:    resp.URL,
		Author: resp.By,
		Score:  resp.Score,
		Time:   time.Unix(resp.Time, 0),
	}
	if len(resp.Kids) > 0 {
		story.CommentCount = len(resp.Kids)
	}
	return story, nil
}
