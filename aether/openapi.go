// aether/openapi.go
//
// Public wrappers around Aether's internal OpenAPI client. These
// wrappers provide a stable, LLM-friendly interface over a variety
// of free, public, no-API-key data sources.

package aether

import (
	"context"
	"fmt"

	"github.com/Nibir1/Aether/internal/model"
)

//
// ────────────────────────────────────────────────
//            PUBLIC DATA MODELS
// ────────────────────────────────────────────────
//

type WikiSummary struct {
	Title       string
	Description string
	Extract     string
	URL         string
	Language    string
}

type HackerNewsStory struct {
	ID           int64
	Title        string
	URL          string
	Author       string
	Score        int
	TimeUnix     int64
	CommentCount int
}

type GitHubReadme struct {
	Owner   string
	Repo    string
	Ref     string
	URL     string
	Content string
}

type WhiteHousePost struct {
	ID       int64
	Title    string
	URL      string
	DateUnix int64
	Excerpt  string
}

type GovernmentPressRelease struct {
	Title    string
	URL      string
	Source   string
	Snippet  string
	DateUnix int64
}

type Weather struct {
	TimeUnix    int64
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	Summary     string
}

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

func (c *Client) WikipediaSummary(ctx context.Context, title string) (*WikiSummary, error) {
	if c == nil || c.openapi == nil {
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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

// HackerNewsTopStoriesDocuments fetches top N stories and converts
// them into model.Document objects ready for JSON / TOON / Lite TOON / BTON pipelines.
func (c *Client) HackerNewsTopStoriesDocuments(ctx context.Context, limit int) ([]*model.Document, error) {
	if c == nil || c.openapi == nil {
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
	}
	return c.openapi.HackerNewsTopStoriesDocuments(ctx, limit)
}

//
// ────────────────────────────────────────────────
//       PUBLIC WRAPPERS — GITHUB README
// ────────────────────────────────────────────────
//

func (c *Client) GitHubReadme(ctx context.Context, owner, repo, ref string) (*GitHubReadme, error) {
	if c == nil || c.openapi == nil {
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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
		return nil, fmt.Errorf("aether: openapi subsystem not initialized")
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
