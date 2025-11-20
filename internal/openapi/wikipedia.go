// internal/openapi/wikipedia.go
//
// Wikipedia integration using the public REST API:
//   https://en.wikipedia.org/api/rest_v1/page/summary/{title}
//
// This endpoint provides a concise, language-agnostic summary that
// is ideal for quick factual lookups and LLM consumption.

package openapi

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// WikiSummary is the internal Wikipedia summary representation.
type WikiSummary struct {
	Title       string
	Description string
	Extract     string
	URL         string
	Language    string
}

// wikipediaSummaryResponse models the subset of the Wikipedia REST
// API summary response that we care about.
type wikipediaSummaryResponse struct {
	Title       string `json:"title"`
	Extract     string `json:"extract"`
	Description string `json:"description"`
	ContentURL  struct {
		Desktop struct {
			Page string `json:"page"`
		} `json:"desktop"`
	} `json:"content_urls"`
	Lang string `json:"lang"`
}

// WikipediaSummary fetches a summary for the given title from the
// English Wikipedia REST API.
//
// Title is automatically URL-escaped and may contain spaces.
func (c *Client) WikipediaSummary(ctx context.Context, title string) (*WikiSummary, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, nil
	}

	escaped := url.PathEscape(title)
	endpoint := "https://en.wikipedia.org/api/rest_v1/page/summary/" + escaped

	body, _, err := c.getJSON(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp wikipediaSummaryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	out := &WikiSummary{
		Title:       resp.Title,
		Description: resp.Description,
		Extract:     resp.Extract,
		URL:         resp.ContentURL.Desktop.Page,
		Language:    resp.Lang,
	}
	return out, nil
}
