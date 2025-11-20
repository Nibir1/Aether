// aether/openapi.go
//
// Public wrappers around Aether's internal OpenAPI client. These
// wrappers provide a stable, LLM-friendly interface over a variety
// of free, public, no-API-key data sources. All integrations are
// implemented internally via shared HTTP client infrastructure.
//
// Supported sources as of Stage 10:
//   - Wikipedia REST API (summaries)
//   - Wikidata (entity lookup + SPARQL)
//   - Hacker News (top stories)
//   - GitHub (README.md)
//   - White House posts (WP-JSON API)
//   - Global government press releases (RSS)
//   - MET Norway weather (global forecast, key-less)

package aether

import (
	"context"
)

//
// ────────────────────────────────────────────────
//            PUBLIC DATA MODELS
// ────────────────────────────────────────────────
//

// WikiSummary is the public Wikipedia summary representation.
type WikiSummary struct {
	Title       string
	Description string
	Extract     string
	URL         string
	Language    string
}

// HackerNewsStory is the public Hacker News story representation.
type HackerNewsStory struct {
	ID           int64
	Title        string
	URL          string
	Author       string
	Score        int
	TimeUnix     int64
	CommentCount int
}

// GitHubReadme is the public GitHub README representation.
type GitHubReadme struct {
	Owner   string
	Repo    string
	Ref     string
	URL     string
	Content string
}

// WhiteHousePost is the public White House post representation.
type WhiteHousePost struct {
	ID       int64
	Title    string
	URL      string
	DateUnix int64
	Excerpt  string
}

// GovernmentPressRelease is the public normalized government press item.
type GovernmentPressRelease struct {
	Title    string
	URL      string
	Source   string
	Snippet  string
	DateUnix int64
}

// Weather represents a normalized hourly weather entry.
type Weather struct {
	TimeUnix    int64
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	Summary     string
}

// WikidataEntity represents a normalized Wikidata entity.
type WikidataEntity struct {
	ID          string
	Title       string
	Description string
	URL         string
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — WIKIPEDIA
// ────────────────────────────────────────────────
//

// WikipediaSummary returns a concise summary for a given topic title
// using the Wikipedia REST API.
func (c *Client) WikipediaSummary(ctx context.Context, title string) (*WikiSummary, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	internal, err := c.openapi.WikipediaSummary(ctx, title)
	if err != nil || internal == nil {
		return nil, err
	}

	return &WikiSummary{
		Title:       internal.Title,
		Description: internal.Description,
		Extract:     internal.Extract,
		URL:         internal.URL,
		Language:    internal.Language,
	}, nil
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — HACKER NEWS
// ────────────────────────────────────────────────
//

func (c *Client) HackerNewsTopStories(ctx context.Context, limit int) ([]HackerNewsStory, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	internalStories, err := c.openapi.HackerNewsTopStories(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]HackerNewsStory, 0, len(internalStories))
	for _, s := range internalStories {
		out = append(out, HackerNewsStory{
			ID:           s.ID,
			Title:        s.Title,
			URL:          s.URL,
			Author:       s.Author,
			Score:        s.Score,
			TimeUnix:     s.Time.Unix(),
			CommentCount: s.CommentCount,
		})
	}
	return out, nil
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — GITHUB README
// ────────────────────────────────────────────────
//

func (c *Client) GitHubReadme(ctx context.Context, owner, repo, ref string) (*GitHubReadme, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	internal, err := c.openapi.GitHubReadme(ctx, owner, repo, ref)
	if err != nil || internal == nil {
		return nil, err
	}

	return &GitHubReadme{
		Owner:   internal.Owner,
		Repo:    internal.Repo,
		Ref:     internal.Ref,
		URL:     internal.URL,
		Content: internal.Content,
	}, nil
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — WHITE HOUSE POSTS
// ────────────────────────────────────────────────
//

func (c *Client) WhiteHouseRecentPosts(ctx context.Context, limit int) ([]WhiteHousePost, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	posts, err := c.openapi.WhiteHouseRecentPosts(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]WhiteHousePost, 0, len(posts))
	for _, p := range posts {
		out = append(out, WhiteHousePost{
			ID:       p.ID,
			Title:    p.Title,
			URL:      p.URL,
			DateUnix: p.Date.Unix(),
			Excerpt:  p.Excerpt,
		})
	}
	return out, nil
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — GOVERNMENT PRESS
// ────────────────────────────────────────────────
//

func (c *Client) GovernmentPress(ctx context.Context, limit int) ([]GovernmentPressRelease, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	items, err := c.openapi.GovernmentPress(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]GovernmentPressRelease, 0, len(items))
	for _, p := range items {
		out = append(out, GovernmentPressRelease{
			Title:    p.Title,
			URL:      p.URL,
			Source:   p.Source,
			Snippet:  p.Snippet,
			DateUnix: p.Date.Unix(),
		})
	}
	return out, nil
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — WEATHER (MET NORWAY)
// ────────────────────────────────────────────────
//

func (c *Client) WeatherAt(ctx context.Context, lat, lon float64, hours int) ([]Weather, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	internal, err := c.openapi.WeatherAt(ctx, lat, lon, hours)
	if err != nil {
		return nil, err
	}

	out := make([]Weather, 0, len(internal))
	for _, w := range internal {
		out = append(out, Weather{
			TimeUnix:    w.Time.Unix(),
			Temperature: w.Temperature,
			Humidity:    w.Humidity,
			WindSpeed:   w.WindSpeed,
			Summary:     w.Summary,
		})
	}
	return out, nil
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — WIKIDATA
// ────────────────────────────────────────────────
//

func (c *Client) WikidataLookup(ctx context.Context, name string) (*WikidataEntity, error) {
	if c == nil || c.openapi == nil {
		return nil, nil
	}

	ent, err := c.openapi.WikidataLookup(ctx, name)
	if err != nil || ent == nil {
		return nil, err
	}

	return &WikidataEntity{
		ID:          ent.ID,
		Title:       ent.Title,
		Description: ent.Description,
		URL:         ent.URL,
	}, nil
}
