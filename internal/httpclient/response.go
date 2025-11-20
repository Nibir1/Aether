// internal/httpclient/response.go
//
// This file defines the internal HTTP response type used by Aether's
// HTTP client. It is deliberately small and immutable from the point of
// view of callers.
package httpclient

import (
	"net/http"
	"time"
)

// Response represents the result of a single HTTP GET operation.
type Response struct {
	URL        string
	StatusCode int
	Header     http.Header
	Body       []byte
	FetchedAt  time.Time
}

// clone creates a deep copy of the response suitable for reuse in
// caches and public-facing results.
func (r *Response) clone() *Response {
	if r == nil {
		return nil
	}
	hdr := make(http.Header, len(r.Header))
	for k, v := range r.Header {
		cp := make([]string, len(v))
		copy(cp, v)
		hdr[k] = cp
	}
	bodyCopy := make([]byte, len(r.Body))
	copy(bodyCopy, r.Body)

	return &Response{
		URL:        r.URL,
		StatusCode: r.StatusCode,
		Header:     hdr,
		Body:       bodyCopy,
		FetchedAt:  r.FetchedAt,
	}
}
