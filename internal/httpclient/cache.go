// internal/httpclient/cache.go
//
// This file implements a simple in-memory cache for HTTP GET responses.
// It is intentionally small and conservative, but provides a real,
// effective caching layer that reduces repeated network requests.
package httpclient

import (
	"sync"
	"time"
)

// memoryCache is a TTL-based in-memory cache keyed by URL string.
type memoryCache struct {
	mu         sync.RWMutex
	entries    map[string]*cacheEntry
	ttl        time.Duration
	maxEntries int
}

type cacheEntry struct {
	resp    *Response
	expires time.Time
}

func newMemoryCache(ttl time.Duration, maxEntries int) *memoryCache {
	if ttl <= 0 {
		return nil
	}
	if maxEntries <= 0 {
		maxEntries = 128
	}
	return &memoryCache{
		entries:    make(map[string]*cacheEntry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}
}

func (c *memoryCache) Get(url string) *Response {
	if c == nil {
		return nil
	}
	now := time.Now()

	c.mu.RLock()
	entry, ok := c.entries[url]
	c.mu.RUnlock()
	if !ok || entry == nil {
		return nil
	}
	if now.After(entry.expires) {
		// expired; evict lazily
		c.mu.Lock()
		delete(c.entries, url)
		c.mu.Unlock()
		return nil
	}
	return entry.resp.clone()
}

func (c *memoryCache) Set(url string, resp *Response) {
	if c == nil || resp == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxEntries {
		// simple eviction: remove an arbitrary entry
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[url] = &cacheEntry{
		resp:    resp.clone(),
		expires: time.Now().Add(c.ttl),
	}
}
