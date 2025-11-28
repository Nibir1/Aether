// aether/cache.go
//
// Public caching configuration options for Aether.
//
// Aether uses a composite caching architecture consisting of:
//   • Memory cache (fast, in-process LRU)
//   • File cache (persistent on disk)
//   • Redis cache (shared/distributed)
//
// The layers can be enabled individually or combined. When multiple layers
// are active, lookups occur in priority order:
//
//     Memory → File → Redis → Miss
//
// Lower-layer hits are automatically promoted upward.
//
// All cache options modify the internal *config.Config before a Client
// instance is created via NewClient.

package aether

import (
	"time"

	icache "github.com/Nibir1/Aether/internal/cache"
	"github.com/Nibir1/Aether/internal/config"
)

// CacheOption mutates internal config during client construction.
type CacheOption func(*config.Config)

//
// ───────────────────────────────────────────────────────────────
//                     MEMORY CACHE CONFIGURATION
// ───────────────────────────────────────────────────────────────
//

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

//
// ───────────────────────────────────────────────────────────────
//                        FILE CACHE CONFIGURATION
// ───────────────────────────────────────────────────────────────
//

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

//
// ───────────────────────────────────────────────────────────────
//                        REDIS CACHE CONFIGURATION
// ───────────────────────────────────────────────────────────────
//

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

//
// ───────────────────────────────────────────────────────────────
//                        CACHE INITIALIZATION
// ───────────────────────────────────────────────────────────────
//

func (c *Client) initCache() {
	if c == nil {
		panic("aether: initCache called on nil *Client")
	}
	if c.cfg == nil {
		panic("aether: initCache requires initialized config")
	}
	if c.logger == nil {
		c.logger = noopLogger{}
	}

	// Normalize negative settings to safe values
	if c.cfg.MaxCacheEntries < 0 {
		c.cfg.MaxCacheEntries = 0
	}

	conf := icache.Config{
		MemoryEnabled: c.cfg.EnableMemoryCache,
		MemoryTTL:     c.cfg.CacheTTL,
		MemoryMax:     c.cfg.MaxCacheEntries,

		FileEnabled:   c.cfg.EnableFileCache,
		FileTTL:       c.cfg.CacheTTL,
		FileDirectory: c.cfg.CacheDirectory,

		RedisEnabled: c.cfg.EnableRedisCache,
		RedisTTL:     c.cfg.CacheTTL,
		RedisAddress: c.cfg.RedisAddress,

		Logger: c.logger,
	}

	// NewComposite returns *Composite (ONE value)
	c.cache = icache.NewComposite(conf)
}

//
// ───────────────────────────────────────────────────────────────
//                         NO-OP LOGGER
// ───────────────────────────────────────────────────────────────
//

// noopLogger satisfies internal/log.Logger while producing no output.
type noopLogger struct{}

func (noopLogger) Infof(string, ...any)  {}
func (noopLogger) Debugf(string, ...any) {}
func (noopLogger) Errorf(string, ...any) {}
func (noopLogger) Warnf(string, ...any)  {} // required by interface
