// aether/aether_fetch.go
//
// Public fetch utilities for Aether's Client.
//
// These helpers expose a safe, legal, robots.txt-compliant interface for
// performing HTTP GET operations using Aether's internal HTTP fetcher.
//
// Plugins MUST use these helpers instead of making their own HTTP requests,
// because these methods ensure:
//   • robots.txt compliance
//   • composite caching (memory + file + redis)
//   • retry logic
//   • rate limiting per-host
//   • unified error handling
//   • automatic gzip/deflate decoding
//
// The internal httpclient.Client performs all legality, safety and caching.
// These public wrappers simply forward to the internal fetcher in a controlled,
// stable way appropriate for plugins and extensions.

package aether

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// FetchRaw performs a robots.txt-compliant HTTP GET and returns:
//
//   - raw response body bytes
//   - http.Header for metadata
//   - error if the fetch fails
//
// The caller receives the untouched response body, suitable for:
//   - JSON decoding
//   - image/binary processing
//   - plugin-level custom parsing
//
// Plugins should ALWAYS use this instead of performing HTTP manually.
func (c *Client) FetchRaw(ctx context.Context, url string) ([]byte, http.Header, error) {
	if c == nil || c.fetcher == nil {
		return nil, nil, fmt.Errorf("aether: client is not initialized")
	}

	resp, err := c.fetcher.Fetch(ctx, url, nil)
	if err != nil {
		return nil, nil, err
	}

	return resp.Body, resp.Header, nil
}

// FetchText performs a robots.txt-compliant fetch and returns the UTF-8
// interpretation of the response body.
//
// This helper is ideal for:
//   - Markdown files
//   - README files
//   - Plain text APIs
//   - HTML extraction by plugins
//
// It returns:
//   - decoded string
//   - http.Header
//   - error
func (c *Client) FetchText(ctx context.Context, url string) (string, http.Header, error) {
	body, hdr, err := c.FetchRaw(ctx, url)
	if err != nil {
		return "", nil, err
	}

	return string(body), hdr, nil
}

// FetchJSON performs a robots.txt-compliant fetch and unmarshals the JSON
// response into the provided destination struct pointer.
//
// Example:
//
//	var out MyStruct
//	if err := cli.FetchJSON(ctx, "https://example.com/api", &out); err != nil {
//	    ...
//	}
//
// This helper is used internally by Aether OpenAPI integrations and is ideal
// for plugins interacting with public JSON APIs.
func (c *Client) FetchJSON(ctx context.Context, url string, dest any) error {
	if dest == nil {
		return fmt.Errorf("aether: FetchJSON requires a non-nil destination")
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
