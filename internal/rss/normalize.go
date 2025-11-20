// internal/rss/normalize.go
//
// Additional helpers for feed normalization. These keep future
// pipeline stages clean.

package rss

import "strings"

// Clean trims and normalizes whitespace in feed fields.
func (f *Feed) Clean() {
	f.Title = strings.TrimSpace(f.Title)
	f.Description = strings.TrimSpace(f.Description)
	for i := range f.Items {
		it := &f.Items[i]
		it.Title = strings.TrimSpace(it.Title)
		it.Description = strings.TrimSpace(it.Description)
		it.Content = strings.TrimSpace(it.Content)
	}
}
