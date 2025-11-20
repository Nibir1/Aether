// plugins/examples/custom_api_plugin/plugin.go
//
// Example: Custom API Source Plugin for Aether
//
// This plugin demonstrates how to implement an Aether SourcePlugin for a
// fictional but realistic public JSON API endpoint.
//
// The intended purpose of this example is to help developers build plugins
// for:
//
//   • Internal company APIs
//   • Public open-data JSON APIs (weather, transport, statistics)
//   • Local microservices
//   • Open educational datasets
//
// The example API behaves like this:
//
//    GET https://api.example.com/v1/info?query=<term>
//
// Returns JSON:
//
//    {
//       "title": "Sample Title",
//       "excerpt": "Short summary",
//       "content": "Long-form article or text",
//       "metadata": {
//           "source": "example",
//           "rating": "A+"
//       },
//       "sections": [
//           {
//              "heading": "Overview",
//              "text": "paragraphs...",
//              "meta": { "version": "1" }
//           }
//       ]
//    }
//
// The plugin fetches this, validates it, and returns a plugins.Document.

package custom_api_plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
)

const apiBase = "https://api.example.com/v1/info"

// apiResponse models the structure returned by the fictional example API.
type apiResponse struct {
	Title    string               `json:"title"`
	Excerpt  string               `json:"excerpt"`
	Content  string               `json:"content"`
	Metadata map[string]string    `json:"metadata"`
	Sections []apiResponseSection `json:"sections"`
}

type apiResponseSection struct {
	Heading string            `json:"heading"`
	Text    string            `json:"text"`
	Meta    map[string]string `json:"meta"`
}

// CustomAPIPlugin is a real demonstration plugin that fetches from
// a generic JSON API.
type CustomAPIPlugin struct {
	client *aether.Client
}

// New creates a new CustomAPIPlugin bound to a specific Aether client.
func New(cli *aether.Client) *CustomAPIPlugin {
	return &CustomAPIPlugin{
		client: cli,
	}
}

func (p *CustomAPIPlugin) Name() string {
	return "custom_api"
}

func (p *CustomAPIPlugin) Description() string {
	return "Example plugin demonstrating how to integrate arbitrary public JSON APIs with Aether."
}

// Capabilities describe the type of queries this plugin may satisfy.
//
// In this example, the plugin tries to answer general factual queries,
// metadata-oriented questions, or anything suitable for a structured API.
func (p *CustomAPIPlugin) Capabilities() []string {
	return []string{"data", "info", "custom", "api"}
}

// Fetch performs the API call and constructs a plugins.Document.
//
// The `query` string is URL-encoded and passed to the example API.
func (p *CustomAPIPlugin) Fetch(ctx context.Context, query string) (*plugins.Document, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("custom_api: empty query is not allowed")
	}

	// Build URL with query parameter
	u, err := url.Parse(apiBase)
	if err != nil {
		return nil, fmt.Errorf("custom_api: invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	// Fetch via Aether's HTTP client (robots + caching + retries)
	body, _, err := p.client.FetchRaw(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("custom_api: HTTP error: %w", err)
	}

	// Decode JSON
	var api apiResponse
	if err := json.Unmarshal(body, &api); err != nil {
		return nil, fmt.Errorf("custom_api: invalid JSON: %w", err)
	}

	// Validate minimal fields
	if strings.TrimSpace(api.Title) == "" {
		return nil, fmt.Errorf("custom_api: missing title in API response")
	}

	// Convert sections
	sections := make([]plugins.Section, 0, len(api.Sections))

	for _, s := range api.Sections {
		role := plugins.SectionRole("section")
		heading := strings.TrimSpace(s.Heading)
		text := strings.TrimSpace(s.Text)

		if heading == "" {
			heading = "Section"
		}

		sections = append(sections, plugins.Section{
			Role:  role,
			Title: heading,
			Text:  text,
			Meta:  s.Meta,
		})
	}

	// Construct plugin Document
	doc := &plugins.Document{
		Source:  "plugin:custom_api",
		URL:     u.String(),
		Kind:    plugins.DocumentKindJSON,
		Title:   api.Title,
		Excerpt: api.Excerpt,
		Content: api.Content,
		Metadata: mergeStringMaps(
			api.Metadata,
			map[string]string{
				"query":        query,
				"fetched_unix": fmt.Sprintf("%d", time.Now().Unix()),
				"plugin":       "custom_api",
			},
		),
		Sections: sections,
	}

	return doc, nil
}

// mergeStringMaps safely merges multiple string maps into a new one.
func mergeStringMaps(ms ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
