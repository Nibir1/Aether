package aether

import (
	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

// MarshalTOONLite serializes SearchResult → normalized → TOON → lite JSON.
func (c *Client) MarshalTOONLite(sr *SearchResult) ([]byte, error) {
	tdoc := c.ToTOON(sr)
	return toon.MarshalLite(tdoc)
}

// MarshalTOONLitePretty pretty prints Lite JSON.
func (c *Client) MarshalTOONLitePretty(sr *SearchResult) ([]byte, error) {
	tdoc := c.ToTOON(sr)
	return toon.MarshalLitePretty(tdoc)
}

// Direct model.Document variant
func (c *Client) MarshalTOONLiteFromModel(doc *model.Document) ([]byte, error) {
	tdoc := c.ToTOONFromModel(doc)
	return toon.MarshalLite(tdoc)
}
