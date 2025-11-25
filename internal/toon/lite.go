package toon

//
// internal/toon/lite.go
//
// TOON-Lite — a minimal JSON representation optimized for:
//   • vector databases
//   • embedding pipelines
//   • compact storage
//
// Reduces: whitespace, empty fields, long field names.
//

import "encoding/json"

type liteToken struct {
	T string            `json:"t"`           // type
	R string            `json:"r,omitempty"` // role
	X string            `json:"x,omitempty"` // text
	A map[string]string `json:"a,omitempty"` // attrs
}

type liteDoc struct {
	U string            `json:"u,omitempty"` // source_url
	K string            `json:"k"`           // kind
	T string            `json:"t,omitempty"` // title
	E string            `json:"e,omitempty"` // excerpt
	A map[string]string `json:"a,omitempty"` // attributes
	N []liteToken       `json:"n"`           // tokens
}

// MarshalLite converts a full TOON Document to a TOON-Lite JSON byte slice.
func MarshalLite(doc *Document) ([]byte, error) {
	out := liteDoc{
		U: doc.SourceURL,
		K: string(doc.Kind),
		T: doc.Title,
		E: doc.Excerpt,
		A: doc.Attributes,
		N: make([]liteToken, len(doc.Tokens)),
	}

	for i, tok := range doc.Tokens {
		out.N[i] = liteToken{
			T: string(tok.Type),
			R: tok.Role,
			X: tok.Text,
			A: tok.Attrs,
		}
	}

	return json.Marshal(out)
}

// MarshalLitePretty pretty-prints the TOON-Lite document.
func MarshalLitePretty(doc *Document) ([]byte, error) {
	out := liteDoc{
		U: doc.SourceURL,
		K: string(doc.Kind),
		T: doc.Title,
		E: doc.Excerpt,
		A: doc.Attributes,
		N: make([]liteToken, len(doc.Tokens)),
	}

	for i, tok := range doc.Tokens {
		out.N[i] = liteToken{
			T: string(tok.Type),
			R: tok.Role,
			X: tok.Text,
			A: tok.Attrs,
		}
	}

	return json.MarshalIndent(out, "", "  ")
}
