// internal/toon/util.go
//
// Utility helpers for TOON documents.

package toon

// ApproxTokenCount returns number of TOON tokens.
// Useful for estimating LLM prompt cost.
func (d *Document) ApproxTokenCount() int {
	if d == nil {
		return 0
	}
	return len(d.Tokens)
}

// TruncateTokens returns a shallow copy of the document with at most N tokens.
func (d *Document) TruncateTokens(n int) *Document {
	if d == nil {
		return nil
	}
	if n <= 0 || len(d.Tokens) <= n {
		return d
	}

	out := *d // shallow copy
	out.Tokens = append([]Token(nil), d.Tokens[:n]...)
	return &out
}

// cloneMap safely clones a map[string]string.
func cloneMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
