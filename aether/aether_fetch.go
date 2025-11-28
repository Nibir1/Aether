// aether/aether_fetch.go
//
// Public fetch utilities for Aether's Client.
//
// These helpers expose a safe, legal, robots.txt-compliant interface for
// performing HTTP GET operations using Aether's internal HTTP fetcher.
//
// Plugins MUST use these helpers instead of making their own HTTP requests,
// because Aether's internal fetcher ensures:
//
//   • robots.txt compliance
//   • composite caching (memory + file + redis)
//   • retry logic (idempotent GET only)
//   • host-scoped rate limiting
//   • timeout + context propagation
//   • gz/deflate automatic decompression
//   • stable error classification
//
// These public wrappers simply forward to the internal fetcher in a controlled,
// stable way appropriate for plugins and extensions.

package aether

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// FetchRaw performs a robots.txt-compliant HTTP GET.
//
// RETURN VALUES:
//
//	body   []byte     — raw response body (unmodified)
//	header http.Header — response headers
//	error
//
// FetchRaw is the lowest-level public API. It does NOT try to interpret the
// response, making it appropriate for:
//
//   - JSON APIs (plugin-side decoding)
//   - images, binaries
//   - HTML/Markdown extraction
//   - RSS/Atom autodetection
//
// NOTE: All legality, robots.txt policy, retries, caching and rate limits are
// enforced inside c.fetcher.Fetch().
func (c *Client) FetchRaw(ctx context.Context, url string) ([]byte, http.Header, error) {
	if c == nil || c.fetcher == nil {
		return nil, nil, fmt.Errorf("aether: client is not initialized")
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return nil, nil, fmt.Errorf("aether: empty URL passed to FetchRaw")
	}

	resp, err := c.fetcher.Fetch(ctx, url, nil)
	if err != nil {
		return nil, nil, err
	}

	return resp.Body, resp.Header, nil
}

// FetchText performs a robots.txt-compliant GET and returns the body as UTF-8.
//
// Ideal for:
//   - README.md
//   - plain-text APIs
//   - HTML before DOM extraction
//   - plugin-readable textual sources
//
// Returns:
//
//	text   string
//	header http.Header
//	error
func (c *Client) FetchText(ctx context.Context, url string) (string, http.Header, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", nil, fmt.Errorf("aether: empty URL passed to FetchText")
	}

	body, hdr, err := c.FetchRaw(ctx, url)
	if err != nil {
		return "", nil, err
	}

	return string(body), hdr, nil
}

// FetchJSON performs a robots.txt-compliant GET and unmarshals JSON.
//
// Example:
//
//	var out MyStruct
//	if err := cli.FetchJSON(ctx, "https://api.example.com", &out); err != nil {
//	    ...
//	}
//
// Errors include network errors, status errors (from FetchRaw), and JSON errors.
func (c *Client) FetchJSON(ctx context.Context, url string, dest any) error {
	if dest == nil {
		return fmt.Errorf("aether: FetchJSON requires a non-nil destination")
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("aether: empty URL passed to FetchJSON")
	}

	body, _, err := c.FetchRaw(ctx, url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("aether: invalid JSON at %s: %w", url, err)
	}

	return nil
}
