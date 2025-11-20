// internal/httpclient/robots.go
//
// This file implements robots.txt fetching and caching for the HTTP
// client. Robots files are fetched once per host and reused.
package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Nibir1/Aether/internal/errors"
	"github.com/Nibir1/Aether/internal/robots"
)

type robotsCache struct {
	mu      sync.Mutex
	entries map[string]*robotsEntry
}

type robotsEntry struct {
	rules *robots.Robots
	// fetchedAt can be used for TTL if desired; for now we keep entries
	// for the lifetime of the process.
	fetchedAt time.Time
}

func newRobotsCache() *robotsCache {
	return &robotsCache{
		entries: make(map[string]*robotsEntry),
	}
}

// allowed checks robots.txt for the given URL and userAgent. It fetches
// and parses robots.txt on first use for each host and caches the result.
//
// If robots.txt cannot be fetched (network error, timeout, 404, etc.),
// Aether treats the host as having no robots rules and allows access.
// This is a common and reasonable default.
func (c *robotsCache) allowed(
	ctx context.Context,
	rawURL string,
	userAgent string,
	client *http.Client,
) (bool, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false, errors.New(errors.KindHTTP, "invalid URL for robots check", err)
	}
	hostKey := parsed.Scheme + "://" + parsed.Host

	entry := c.get(hostKey)
	if entry == nil {
		entry, err = c.fetch(ctx, hostKey, client)
		if err != nil {
			// On robots fetch failure, allow by default but surface the
			// error to the caller via log if needed.
			return true, nil
		}
	}

	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	return entry.rules.Allowed(userAgent, path), nil
}

func (c *robotsCache) get(hostKey string) *robotsEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.entries[hostKey]
}

func (c *robotsCache) fetch(
	ctx context.Context,
	hostKey string,
	client *http.Client,
) (*robotsEntry, error) {
	c.mu.Lock()
	if entry, ok := c.entries[hostKey]; ok {
		c.mu.Unlock()
		return entry, nil
	}
	c.mu.Unlock()

	robotsURL := fmt.Sprintf("%s/robots.txt", hostKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return nil, errors.New(errors.KindHTTP, "creating robots.txt request failed", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(errors.KindHTTP, "fetching robots.txt failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// treat as no rules
		entry := &robotsEntry{
			rules:     &robots.Robots{},
			fetchedAt: time.Now(),
		}
		c.mu.Lock()
		c.entries[hostKey] = entry
		c.mu.Unlock()
		return entry, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(errors.KindHTTP, "reading robots.txt failed", err)
	}

	rules := robots.Parse(body)
	entry := &robotsEntry{
		rules:     rules,
		fetchedAt: time.Now(),
	}

	c.mu.Lock()
	c.entries[hostKey] = entry
	c.mu.Unlock()

	return entry, nil
}
