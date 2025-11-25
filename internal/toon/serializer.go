// internal/toon/serializer.go
//
// JSON serialization helpers for TOON documents.
//
// These helpers wrap encoding/json to provide a single place for
// future canonicalization rules (e.g., versioning, field ordering)
// without forcing callers to re-implement serialization logic.

package toon

import (
	"encoding/json"
)

// MarshalJSONCompact serializes the TOON document into compact JSON.
// This is intended for storage or machine-to-machine usage.
func (d *Document) MarshalJSONCompact() ([]byte, error) {
	if d == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(d)
}

// MarshalJSONPretty serializes the TOON document into pretty-printed JSON.
// This is ideal for debugging, logs, and human inspection.
func (d *Document) MarshalJSONPretty() ([]byte, error) {
	if d == nil {
		return json.MarshalIndent(nil, "", "  ")
	}
	return json.MarshalIndent(d, "", "  ")
}
