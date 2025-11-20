// plugins/examples/hn_plugin/hn.go
//
// Example: Hacker News Source Plugin for Aether
//
// This plugin demonstrates how to implement a SourcePlugin that fetches
// legal, public data from the Hacker News Firebase API. The API is fully
// public, requires no API keys, has no authentication, and explicitly
// allows programmatic access.
//
// The plugin fetches:
//   • Top story IDs
//   • Individual story metadata
//
// and returns a plugins.Document containing:
//   • Title
//   • Excerpt
//   • Metadata (score, author, URL, etc.)
//   • Feed-style sections for each story
//
// This plugin intentionally does NOT perform any direct HTTP calls.
// Instead, it delegates to Aether’s Search/OpenAPI infrastructure,
// as plugin authors SHOULD NOT bypass robots.txt or other restrictions.

package hn_plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
)

// hnAPI is the official Firebase endpoint.
const hnAPIBase = "https://hacker-news.firebaseio.com/v0"

// hnStory represents the JSON returned by the HN API.
type hnStory struct {
	ID    int64   `json:"id"`
	Title string  `json:"title"`
	URL   string  `json:"url"`
	By    string  `json:"by"`
	Score int     `json:"score"`
	Time  int64   `json:"time"`
	Kids  []int64 `json:"kids"`
	Type  string  `json:"type"`
}

// HNPlugin implements a basic Hacker News SourcePlugin.
type HNPlugin struct {
	client *aether.Client
	limit  int
}

// New creates a new HN plugin. The limit determines how many top stories
// should be fetched (e.g., 5, 10, 20).
func New(cli *aether.Client, limit int) *HNPlugin {
	if limit <= 0 {
		limit = 10
	}
	return &HNPlugin{
		client: cli,
		limit:  limit,
	}
}

// Name returns the unique plugin identifier.
func (p *HNPlugin) Name() string {
	return "hackernews"
}

// Description returns a human-friendly summary of what this plugin does.
func (p *HNPlugin) Description() string {
	return "Fetches top Hacker News stories using the public HN Firebase API."
}

// Capabilities describe what queries this plugin is suitable for.
// SmartQuery routing can use these to match user intent.
func (p *HNPlugin) Capabilities() []string {
	return []string{"news", "tech", "hn", "hackernews"}
}

// Fetch retrieves the top Hacker News stories.
// The query argument is ignored (this plugin is not query-sensitive).
func (p *HNPlugin) Fetch(ctx context.Context, query string) (*plugins.Document, error) {
	// 1. Fetch top story IDs
	topURL := hnAPIBase + "/topstories.json"
	body, _, err := p.client.FetchRaw(ctx, topURL) // we will add FetchRaw publicly later
	if err != nil {
		return nil, fmt.Errorf("hn: failed to fetch topstories: %w", err)
	}

	var ids []int64
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, fmt.Errorf("hn: invalid topstories JSON: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("hn: no top stories returned")
	}

	// Limit
	if len(ids) > p.limit {
		ids = ids[:p.limit]
	}

	// 2. Fetch each story
	sections := make([]plugins.Section, 0, len(ids))

	for _, id := range ids {
		story, err := p.fetchStory(ctx, id)
		if err != nil {
			continue // skip failures safely
		}

		sections = append(sections, plugins.Section{
			Role:  plugins.SectionRole("feed_item"),
			Title: story.Title,
			Text:  p.buildStorySnippet(story),
			Meta: map[string]string{
				"url":        story.URL,
				"score":      strconv.Itoa(story.Score),
				"author":     story.By,
				"timestamp":  strconv.FormatInt(story.Time, 10),
				"story_type": story.Type,
			},
		})
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("hn: no valid stories retrieved")
	}

	// 3. Build final plugin Document
	doc := &plugins.Document{
		Source:  "plugin:hackernews",
		URL:     "https://news.ycombinator.com/",
		Kind:    plugins.DocumentKindFeed,
		Title:   "Hacker News — Top Stories",
		Excerpt: "Top stories from Hacker News, powered by the public Firebase API.",
		Content: "",
		Metadata: map[string]string{
			"fetch_time_unix": strconv.FormatInt(time.Now().Unix(), 10),
		},
		Sections: sections,
	}

	return doc, nil
}

// fetchStory loads the JSON for a single HN story.
func (p *HNPlugin) fetchStory(ctx context.Context, id int64) (*hnStory, error) {
	url := fmt.Sprintf("%s/item/%d.json", hnAPIBase, id)

	body, _, err := p.client.FetchRaw(ctx, url)
	if err != nil {
		return nil, err
	}

	var story hnStory
	if err := json.Unmarshal(body, &story); err != nil {
		return nil, err
	}

	return &story, nil
}

// buildStorySnippet returns a compact, readable preview of a story.
func (p *HNPlugin) buildStorySnippet(s *hnStory) string {
	var b strings.Builder

	b.WriteString(s.Title)

	if s.URL != "" {
		b.WriteString(" — ")
		b.WriteString(s.URL)
	}

	b.WriteString("\nBy ")
	b.WriteString(s.By)
	b.WriteString(" | Score: ")
	b.WriteString(strconv.Itoa(s.Score))

	return b.String()
}
