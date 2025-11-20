// internal/openapi/wikidata.go
//
// Wikidata integration using:
//   - Special:EntityData/{id}.json (no key required)
//   - Public SPARQL endpoint (legal & free)
//
// Wikidata is perfect for structured facts, IDs, names, descriptions,
// and relationships.

package openapi

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// WikidataEntity is the normalized Wikidata entity info.
type WikidataEntity struct {
	ID          string
	Title       string
	Description string
	URL         string
}

// WikidataLookup searches Wikidata for the best entity match using SPARQL.
func (c *Client) WikidataLookup(ctx context.Context, name string) (*WikidataEntity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}

	sparql := `
SELECT ?item ?itemLabel WHERE {
  ?item rdfs:label "` + escape(name) + `"@en.
}
LIMIT 1
`

	queryURL := "https://query.wikidata.org/sparql?format=json&query=" + url.QueryEscape(sparql)

	body, _, err := c.getJSON(ctx, queryURL)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Results struct {
			Bindings []struct {
				Item struct {
					Value string `json:"value"`
				} `json:"item"`
				Label struct {
					Value string `json:"value"`
				} `json:"itemLabel"`
			} `json:"bindings"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Results.Bindings) == 0 {
		return nil, nil
	}

	item := resp.Results.Bindings[0].Item.Value
	parts := strings.Split(item, "/")
	id := parts[len(parts)-1]

	// Fetch full entity data
	entityURL := "https://www.wikidata.org/wiki/Special:EntityData/" + id + ".json"

	raw, _, err := c.getJSON(ctx, entityURL)
	if err != nil {
		return nil, err
	}

	var ent struct {
		Entities map[string]struct {
			Labels       map[string]struct{ Value string } `json:"labels"`
			Descriptions map[string]struct{ Value string } `json:"descriptions"`
		} `json:"entities"`
	}

	if err := json.Unmarshal(raw, &ent); err != nil {
		return nil, err
	}

	itemData := ent.Entities[id]

	out := &WikidataEntity{
		ID:  id,
		URL: "https://www.wikidata.org/wiki/" + id,
	}

	if v, ok := itemData.Labels["en"]; ok {
		out.Title = v.Value
	}
	if v, ok := itemData.Descriptions["en"]; ok {
		out.Description = v.Value
	}

	return out, nil
}

func escape(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
