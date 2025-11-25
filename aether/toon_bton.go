package aether

import (
	"github.com/Nibir1/Aether/internal/model"
	"github.com/Nibir1/Aether/internal/toon"
)

// MarshalBTON serializes a SearchResult into BT0N binary format.
func (c *Client) MarshalBTON(sr *SearchResult) ([]byte, error) {
	doc := c.ToTOON(sr)
	return toon.EncodeBTON(doc)
}

// UnmarshalBTON parses BT0N back into a TOON Document.
func (c *Client) UnmarshalBTON(b []byte) (*toon.Document, error) {
	return toon.DecodeBTON(b)
}

// MarshalBTONFromModel serializes a model.Document into BT0N.
func (c *Client) MarshalBTONFromModel(doc *model.Document) ([]byte, error) {
	tdoc := c.ToTOONFromModel(doc)
	return toon.EncodeBTON(tdoc)
}
