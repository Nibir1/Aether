// aether/fetch.go
//
// This file exposes the Fetch method on Client, which performs a
// robots.txt-compliant HTTP GET using Aether's internal HTTP client.
package aether

import (
	"context"
	"net/http"
	"time"

	hclient "github.com/Nibir1/Aether/internal/httpclient"
)

// FetchResult is the public view of a completed HTTP fetch operation.
//
// It intentionally exposes only safe, immutable data. Callers should
// treat the Body slice as read-only.
type FetchResult struct {
	URL        string
	StatusCode int
	Header     http.Header
	Body       []byte
	FetchedAt  time.Time
}

// FetchOptions describes optional parameters for Fetch.
type FetchOptions struct {
	Headers http.Header
}

// FetchOption configures FetchOptions.
type FetchOption func(*FetchOptions)

// WithHeader adds or overrides a single HTTP header for the Fetch call.
//
// Multiple WithHeader options can be combined; later calls override
// earlier ones for the same header key.
func WithHeader(key, value string) FetchOption {
	return func(o *FetchOptions) {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}
		o.Headers.Set(key, value)
	}
}

// Fetch performs a robots.txt-compliant HTTP GET for the given URL.
//
// It automatically:
//   - respects robots.txt rules using the configured User-Agent,
//   - applies polite per-host concurrency limits,
//   - performs simple in-memory response caching,
//   - retries transient failures with exponential backoff.
//
// The result is a FetchResult containing status, headers, body and
// timestamp of retrieval.
func (c *Client) Fetch(ctx context.Context, rawURL string, opts ...FetchOption) (*FetchResult, error) {
	if c == nil || c.fetcher == nil {
		return nil, &Error{
			Kind: ErrorKindConfig,
			Msg:  "Aether client is not initialized",
		}
	}

	var fo FetchOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&fo)
		}
	}

	resp, err := c.fetcher.Fetch(ctx, rawURL, fo.Headers)
	if err != nil {
		return nil, err
	}

	return &FetchResult{
		URL:        resp.URL,
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       resp.Body, // treat as read-only
		FetchedAt:  resp.FetchedAt,
	}, nil
}

// internal helper to adapt errors when needed.
// Currently this just uses the internal type directly, but keeping
// the function around makes it easy to adjust error mapping later.
func adaptHTTPError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*hclient.Error); ok {
		// httpclient.Error is already a structured error type
		// compatible with Aether's public error alias.
		return err
	}
	return err
}
