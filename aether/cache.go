// aether/cache.go
//
// This file defines the public caching configuration options for Aether.
// The caching subsystem is a core performance, efficiency, and politeness
// mechanism: it reduces redundant network calls, minimizes bandwidth usage,
// speeds up high-level operations (Search, RSS, OpenAPIs), and enforces
// predictable TTL behaviors.
//
// Aether supports a composite cache architecture with up to three layers:
//
//   1. Memory (fast LRU in-memory)
//   2. File (hashed filenames, persistent between runs)
//   3. Redis (shared/distributed)
//
// Each layer may be enabled independently. When multiple layers are enabled,
// the composite cache checks them in priority order:
//
//   Memory → File → Redis → Miss
//
// Cache hits at lower layers automatically promote the entry upward.

package aether

import (
	"time"

	icache "github.com/Nibir1/Aether/internal/cache"
	"github.com/Nibir1/Aether/internal/config"
)

// CacheOption configures Aether caching behavior.
//
// These options modify the internal *config.Config used by the Client.
// They must be applied when constructing the Aether Client via NewClient.
type CacheOption func(*config.Config)

// WithMemoryCache enables the in-memory LRU cache.
//
// maxEntries sets the maximum number of items retained in memory.
// ttl sets the time-to-live for cached entries.
func WithMemoryCache(maxEntries int, ttl time.Duration) CacheOption {
	return func(c *config.Config) {
		if maxEntries > 0 {
			c.MaxCacheEntries = maxEntries
		}
		if ttl > 0 {
			c.CacheTTL = ttl
		}
		c.EnableMemoryCache = true
	}
}

// WithFileCache enables the persistent file-backed cache.
//
// dir must be a writable directory for Aether to store its cached files.
// ttl defines how long entries remain valid.
func WithFileCache(dir string, ttl time.Duration) CacheOption {
	return func(c *config.Config) {
		if dir != "" {
			c.CacheDirectory = dir
			c.EnableFileCache = true
		}
		if ttl > 0 {
			c.CacheTTL = ttl
		}
	}
}

// WithRedisCache enables Redis caching.
//
// addr is the Redis server address (e.g., "localhost:6379").
// ttl controls how long redis entries remain valid.
func WithRedisCache(addr string, ttl time.Duration) CacheOption {
	return func(c *config.Config) {
		if addr != "" {
			c.RedisAddress = addr
			c.EnableRedisCache = true
		}
		if ttl > 0 {
			c.CacheTTL = ttl
		}
	}
}

// initCache attaches the unified composite cache to the Client.
//
// This function is invoked internally by NewClient after all configuration
// options have been applied. It builds a composite cache using the settings
// stored in the internal configuration.
func (c *Client) initCache() {
	c.cache = icache.NewComposite(icache.Config{

		// Memory cache layer
		MemoryEnabled: c.cfg.EnableMemoryCache,
		MemoryTTL:     c.cfg.CacheTTL,
		MemoryMax:     c.cfg.MaxCacheEntries,

		// File cache layer
		FileEnabled:   c.cfg.EnableFileCache,
		FileTTL:       c.cfg.CacheTTL,
		FileDirectory: c.cfg.CacheDirectory,

		// Redis cache layer
		RedisEnabled: c.cfg.EnableRedisCache,
		RedisTTL:     c.cfg.CacheTTL,
		RedisAddress: c.cfg.RedisAddress,

		// Logging integration
		Logger: c.logger,
	})
}
