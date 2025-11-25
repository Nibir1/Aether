// internal/toon/lite_test.go
package toon

import (
	"encoding/json"
	"testing"

	"github.com/Nibir1/Aether/internal/model"
)

func TestMarshalLite_BasicStructure(t *testing.T) {
	m := &model.Document{
		SourceURL: "https://example.com/entity",
		Kind:      model.DocumentKindEntity,
		Title:     "Sample Entity",
		Excerpt:   "Entity summary",
		Metadata: map[string]string{
			"id": "Q12345",
		},
		Sections: []model.Section{
			{
				Role:    model.SectionRoleEntity,
				Heading: "Properties",
				Text:    "Key/value data here",
				Meta: map[string]string{
					"prop_count": "1",
				},
			},
		},
	}

	tdoc := FromModel(m)

	data, err := MarshalLite(tdoc)
	if err != nil {
		t.Fatalf("MarshalLite error: %v", err)
	}

	var decoded liteDoc
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal liteDoc error: %v", err)
	}

	if decoded.K != string(m.Kind) {
		t.Fatalf("kind mismatch: got %q, want %q", decoded.K, m.Kind)
	}
	if decoded.U != m.SourceURL {
		t.Fatalf("source URL mismatch: got %q, want %q", decoded.U, m.SourceURL)
	}
	if len(decoded.N) == 0 {
		t.Fatal("expected at least one token in liteDoc.N")
	}
}
