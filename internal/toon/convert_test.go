// internal/toon/convert_test.go
package toon

import (
	"testing"

	"github.com/Nibir1/Aether/internal/model"
)

func TestFromModel_BasicFields(t *testing.T) {
	m := &model.Document{
		SourceURL: "https://example.com/article",
		Kind:      model.DocumentKindArticle,
		Title:     "Hello World",
		Excerpt:   "This is an excerpt.",
		Content:   "This is the main body content.",
		Metadata: map[string]string{
			"aether.intent": "article",
			"lang":          "en",
		},
		Sections: []model.Section{
			{
				Role:    model.SectionRoleBody,
				Heading: "Introduction",
				Text:    "Intro body text.",
				Meta: map[string]string{
					"author": "alice",
				},
			},
		},
	}

	tdoc := FromModel(m)

	if tdoc.SourceURL != m.SourceURL {
		t.Fatalf("SourceURL mismatch: got %q, want %q", tdoc.SourceURL, m.SourceURL)
	}
	if tdoc.Kind != m.Kind {
		t.Fatalf("Kind mismatch: got %q, want %q", tdoc.Kind, m.Kind)
	}
	if tdoc.Title != m.Title {
		t.Fatalf("Title mismatch: got %q, want %q", tdoc.Title, m.Title)
	}
	if tdoc.Excerpt != m.Excerpt {
		t.Fatalf("Excerpt mismatch: got %q, want %q", tdoc.Excerpt, m.Excerpt)
	}

	if len(tdoc.Tokens) == 0 {
		t.Fatal("expected non-empty token stream from FromModel")
	}

	if tdoc.ApproxTokenCount() != len(tdoc.Tokens) {
		t.Fatalf("ApproxTokenCount mismatch: got %d, want %d", tdoc.ApproxTokenCount(), len(tdoc.Tokens))
	}
}
