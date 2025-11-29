// internal/openapi/hackernews.go
//
// Hacker News integration using the official Firebase API:
//   https://hacker-news.firebaseio.com/v0/topstories.json
//   https://hacker-news.firebaseio.com/v0/item/{id}.json
//
// Fully legal, JSON-based API access. Provides normalization into
// model.Document for JSON / TOON / Lite TOON / BTON pipelines.

package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Nibir1/Aether/internal/errors"
	"github.com/Nibir1/Aether/internal/model"
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
// limit is clamped to 50 to avoid excessive calls.
func (c *Client) HackerNewsTopStories(ctx context.Context, limit int) ([]HNStory, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	body, _, err := c.getJSON(ctx, "https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		return nil, errors.New(errors.KindHTTP, "failed to fetch topstories.json", err)
	}

	var ids []int64
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, errors.New(errors.KindParsing, "failed to parse topstories.json", err)
	}

	if len(ids) < limit {
		limit = len(ids)
	}

	// Concurrent fetch
	type result struct {
		story *HNStory
	}
	outCh := make(chan result, limit)
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, 5) // max 5 concurrent requests

	for i := 0; i < limit; i++ {
		id := ids[i]
		wg.Add(1)

		go func(itemID int64) {
			defer wg.Done()
			sem <- struct{}{}
			story, err := c.hnFetchItem(ctx, itemID)
			<-sem

			if err != nil {
				log.Printf("[HN] failed to fetch item %d: %v", itemID, err)
			}
			if story != nil {
				outCh <- result{story: story}
			}
		}(id)
	}

	wg.Wait()
	close(outCh)

	stories := make([]HNStory, 0, limit)
	for r := range outCh {
		if r.story != nil {
			stories = append(stories, *r.story)
		}
	}

	return stories, nil
}

// HackerNewsTopStoriesDocuments fetches top N stories and converts them
// into model.Document objects ready for JSON / TOON pipelines.
func (c *Client) HackerNewsTopStoriesDocuments(ctx context.Context, limit int) ([]*model.Document, error) {
	hnStories, err := c.HackerNewsTopStories(ctx, limit)
	if err != nil {
		return nil, err
	}

	docs := make([]*model.Document, 0, len(hnStories))
	for _, s := range hnStories {
		d := &model.Document{
			SourceURL: s.URL,
			Kind:      model.DocumentKindArticle,
			Title:     s.Title,
			Excerpt:   fmt.Sprintf("HN story by %s, score %d, comments %d", s.Author, s.Score, s.CommentCount),
			Content: fmt.Sprintf("Title: %s\nAuthor: %s\nScore: %d\nComments: %d\nURL: %s\n",
				s.Title, s.Author, s.Score, s.CommentCount, s.URL),
			Metadata: map[string]string{
				"source": "hackernews",
				"hn.id":  fmt.Sprintf("%d", s.ID),
			},
		}
		docs = append(docs, d)
	}

	return docs, nil
}

// hnFetchItem fetches and normalizes a single Hacker News item.
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

	// Only keep "story" items
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
