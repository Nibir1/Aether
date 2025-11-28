// internal/httpclient/robots.go
//
// This file implements robots.txt fetching and caching for the HTTP
// client. Robots files are fetched once per host and reused.
//
// Option A (Host-Level Override):
// --------------------------------
// Aether remains 100% compliant by default. If the user explicitly
// enables RobotsOverrideEnabled and provides RobotsAllowedHosts,
// Aether will *skip* robots.txt checks ONLY for those hosts.
//
// This preserves legal safety: the user must opt-in to override.
//

package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Nibir1/Aether/internal/config"
	"github.com/Nibir1/Aether/internal/errors"
	"github.com/Nibir1/Aether/internal/robots"
)

type robotsCache struct {
	mu      sync.Mutex
	entries map[string]*robotsEntry
	cfg     *config.Config
}

type robotsEntry struct {
	rules     *robots.Robots
	fetchedAt time.Time
}

// newRobotsCache creates a robots cache with reference to global config.
func newRobotsCache(cfg *config.Config) *robotsCache {
	return &robotsCache{
		entries: make(map[string]*robotsEntry),
		cfg:     cfg,
	}
}

// canonicalHost normalizes hosts for lookup:
// - lowercase
// - remove :port
func canonicalHost(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	if h == "" {
		return ""
	}
	// strip port
	if idx := strings.IndexByte(h, ':'); idx != -1 {
		h = h[:idx]
	}
	return h
}

// allowed checks whether access to rawURL is permitted by robots.txt.
//
// Behavior:
// ---------
// 1. If override mode is ON and host is allowed → return true immediately.
// 2. Otherwise, perform standard robots.txt fetch + cache.
// 3. If robots.txt cannot be fetched → allow by default.
// 4. Else check Robots rules.
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

	hostName := canonicalHost(parsed.Host)
	hostKey := parsed.Scheme + "://" + parsed.Host

	//
	// ──────────────────────────────────────────────
	// OPTION A: Host-level robots override
	// ──────────────────────────────────────────────
	//
	if c.cfg.RobotsOverrideEnabled && hostName != "" {
		for _, allowed := range c.cfg.RobotsAllowedHosts {
			if canonicalHost(allowed) == hostName {
				// User explicitly granted override permission.
				return true, nil
			}
		}
	}

	//
	// ──────────────────────────────────────────────
	// Standard robots.txt handling
	// ──────────────────────────────────────────────
	//

	entry := c.get(hostKey)
	if entry == nil {
		entry, err = c.fetch(ctx, hostKey, client)
		if err != nil {
			// Fail open: if robots cannot be fetched, allow access.
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

	// If robots.txt missing → treat as empty rules
	if resp.StatusCode >= 400 {
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
