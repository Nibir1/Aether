// internal/toon/bton_test.go
package toon

import (
	"testing"

	"github.com/Nibir1/Aether/internal/model"
)

func TestBTON_RoundTrip(t *testing.T) {
	m := &model.Document{
		SourceURL: "https://example.com/feed",
		Kind:      model.DocumentKindFeed,
		Title:     "Feed Title",
		Excerpt:   "Short summary",
		Metadata: map[string]string{
			"lang": "en",
		},
		Sections: []model.Section{
			{
				Role:    model.SectionRoleFeedItem,
				Heading: "Item 1",
				Text:    "Item body content",
				Meta: map[string]string{
					"link": "https://example.com/item1",
				},
			},
		},
	}

	tdoc := FromModel(m)
	b, err := EncodeBTON(tdoc)
	if err != nil {
		t.Fatalf("EncodeBTON error: %v", err)
	}

	out, err := DecodeBTON(b)
	if err != nil {
		t.Fatalf("DecodeBTON error: %v", err)
	}

	if out.Kind != tdoc.Kind {
		t.Fatalf("Kind mismatch after round-trip: got %q, want %q", out.Kind, tdoc.Kind)
	}
	if out.Title != tdoc.Title {
		t.Fatalf("Title mismatch after round-trip: got %q, want %q", out.Title, tdoc.Title)
	}
	if len(out.Tokens) != len(tdoc.Tokens) {
		t.Fatalf("Token count mismatch after round-trip: got %d, want %d", len(out.Tokens), len(tdoc.Tokens))
	}
}
