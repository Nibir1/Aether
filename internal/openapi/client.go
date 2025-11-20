// internal/openapi/client.go
//
// Package openapi implements Aether's integrations with public,
// legal, API-key-free HTTP data sources, including:
//
//   - Wikipedia REST API
//   - Wikidata Entity & SPARQL API
//   - Hacker News Firebase API
//   - GitHub Raw Content API (README)
//   - White House WP-JSON API
//   - Government RSS/Atom press feeds
//   - MET Norway weather API
//
// All outbound requests pass through Aether’s unified HTTP client,
// which applies timeouts, caching, polite-concurrency, logging,
// and robots.txt-compliant behavior (where applicable).
//
// This file defines the shared OpenAPI client infrastructure that
// all individual service modules rely upon.

package openapi

import (
	"context"
	"net/http"

	"github.com/Nibir1/Aether/internal/config"
	"github.com/Nibir1/Aether/internal/httpclient"
	"github.com/Nibir1/Aether/internal/log"
)

// Client is Aether's internal aggregator for all OpenAPI integrations.
//
// The Client is intentionally simple: it wraps the shared internal HTTP
// fetcher. Actual service logic (Wikipedia, HN, GitHub, etc.) lives in
// separate files within this package.
//
// The design allows:
//   - zero-copy reuse of Aether’s HTTP/cache/robots pipeline
//   - strong isolation between integrations
//   - safe concurrent usage across goroutines
//
// OpenAPI Client is safe for concurrent use.
type Client struct {
	cfg    *config.Config
	logger log.Logger
	http   *httpclient.Client
}

// New constructs a new OpenAPI client from the shared internal HTTP client.
//
// This constructor is invoked by aether.NewClient and must receive the same
// httpclient.Client instance that the rest of Aether uses to ensure consistent
// caching, User-Agent identity, and robots.txt behavior.
//
// cfg and logger are passed through without modification so that all OpenAPI
// integrations use the same configuration and logging preferences.
func New(cfg *config.Config, logger log.Logger, httpClient *httpclient.Client) *Client {
	return &Client{
		cfg:    cfg,
		logger: logger,
		http:   httpClient,
	}
}

// getJSON executes a GET request through Aether’s fetch pipeline and returns
// the response body and headers for JSON decoding.
//
// The caller is responsible for unmarshalling the JSON.
func (c *Client) getJSON(ctx context.Context, url string) ([]byte, http.Header, error) {
	resp, err := c.http.Fetch(ctx, url, nil)
	if err != nil {
		return nil, nil, err
	}
	return resp.Body, resp.Header, nil
}

// getText executes a GET request intended for plain-text sources such as
// GitHub README documents, plain RSS feeds, or metadata endpoints.
//
// The caller receives the raw bytes and HTTP headers.
func (c *Client) getText(ctx context.Context, url string) ([]byte, http.Header, error) {
	resp, err := c.http.Fetch(ctx, url, nil)
	if err != nil {
		return nil, nil, err
	}
	return resp.Body, resp.Header, nil
}

// getXML is a small helper used by RSS/Atom and XML-structured APIs.
// It does not interpret charset conversion automatically — that is handled
// by the individual XML integration files where needed.
func (c *Client) getXML(ctx context.Context, url string) ([]byte, http.Header, error) {
	resp, err := c.http.Fetch(ctx, url, nil)
	if err != nil {
		return nil, nil, err
	}
	return resp.Body, resp.Header, nil
}
