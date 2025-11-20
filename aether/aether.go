// aether/aether.go
//
// Package aether provides the public entrypoint for the Aether library.
// Aether is a legal, robots.txt-compliant web retrieval toolkit that
// normalizes public, open data into JSON- and TOON-compatible models.
//
// As of Stage 10, Aether includes:
//   - unified composite caching (memory, file, redis)
//   - robots.txt-compliant HTTP fetcher
//   - HTML parsing + article extraction
//   - Detect/Meta subsystem
//   - SmartQuery router
//   - RSS/Atom subsystem
//   - OpenAPI integrations (Wikipedia, Wikidata, HN, GitHub, GovPress,
//     WhiteHouse, Weather via MET Norway)

package aether

import (
	"fmt"
	"time"

	icache "github.com/Nibir1/Aether/internal/cache"
	"github.com/Nibir1/Aether/internal/config"
	hclient "github.com/Nibir1/Aether/internal/httpclient"
	"github.com/Nibir1/Aether/internal/log"
	iopenapi "github.com/Nibir1/Aether/internal/openapi"
	"github.com/Nibir1/Aether/internal/version"
)

// DefaultUserAgent is the default HTTP User-Agent string that Aether uses
// when making outbound requests. It includes the repository for transparency.
const DefaultUserAgent = "AetherBot/1.0 (+https://github.com/Nibir1/Aether)"

// Client is the main public interface for using Aether.
//
// The Client owns:
//   - internal unified cache (memory + file + redis)
//   - internal robots.txt-compliant HTTP fetcher
//   - internal OpenAPI aggregation client
//
// In later stages it will also expose:
//   - TOON/JSON serialization
//   - full multi-source query federation
//   - Crawl, Batch, Normalize, Display subsystems
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	fetcher *hclient.Client
	cache   icache.Cache
	openapi *iopenapi.Client
}

// Config is the public, inspectable view of effective Aether configuration.
//
// This is intentionally separate from internal config.Config to ensure internal
// changes do not break the public API surface.
type Config struct {
	// Networking
	UserAgent          string
	RequestTimeout     time.Duration
	MaxConcurrentHosts int
	MaxRequestsPerHost int

	// Logging
	EnableDebugLogging bool

	// Caching
	EnableMemoryCache bool
	EnableFileCache   bool
	EnableRedisCache  bool

	CacheDirectory string
	RedisAddress   string

	CacheTTL        time.Duration
	MaxCacheEntries int
}

// Option is a functional option that modifies the internal configuration.
//
// Aether uses this pattern to keep NewClient future-proof as the library gains
// new capabilities and settings.
type Option func(*config.Config)

// NewClient constructs a new Aether Client with optional configuration.
//
// Pipeline:
//  1. Load default internal config
//  2. Apply user-specified Option values
//  3. Ensure a User-Agent is set
//  4. Initialize logger
//  5. Initialize unified composite cache
//  6. Initialize HTTP fetcher with robots.txt and cache support
//  7. Initialize internal OpenAPI client (Wikipedia, HN, GitHub, GovPress…)
func NewClient(opts ...Option) (*Client, error) {
	internalCfg := config.Default()

	for _, opt := range opts {
		if opt != nil {
			opt(internalCfg)
		}
	}

	// Default UA if caller did not specify one.
	if internalCfg.UserAgent == "" {
		internalCfg.UserAgent = DefaultUserAgent
	}

	logger := log.New(internalCfg.EnableDebugLogging)

	cli := &Client{
		cfg:    internalCfg,
		logger: logger,
	}

	// (5) unified composite cache
	cli.initCache()

	// (6) robots.txt-compliant HTTP fetcher wired to composite cache
	cli.fetcher = hclient.New(internalCfg, logger, cli.cache)

	// (7) OpenAPI client (no API keys; all public sources)
	cli.openapi = iopenapi.New(internalCfg, logger, cli.fetcher)

	return cli, nil
}

//
// ────────────────────────────────────────────────
//      PUBLIC CONFIGURATION OPTIONS
// ────────────────────────────────────────────────
//

// WithUserAgent overrides the default HTTP User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *config.Config) {
		if ua != "" {
			c.UserAgent = ua
		}
	}
}

// WithRequestTimeout sets the maximum duration Aether will wait on HTTP GET.
func WithRequestTimeout(d time.Duration) Option {
	return func(c *config.Config) {
		if d > 0 {
			c.RequestTimeout = d
		}
	}
}

// WithConcurrency sets concurrency caps for outbound HTTP I/O.
func WithConcurrency(maxHosts, maxPerHost int) Option {
	return func(c *config.Config) {
		if maxHosts > 0 {
			c.MaxConcurrentHosts = maxHosts
		}
		if maxPerHost > 0 {
			c.MaxRequestsPerHost = maxPerHost
		}
	}
}

// WithDebugLogging enables verbose internal logs.
// Useful for debugging; disabled by default in production.
func WithDebugLogging(enabled bool) Option {
	return func(c *config.Config) {
		c.EnableDebugLogging = enabled
	}
}

//
// ────────────────────────────────────────────────
//              PUBLIC UTILITIES
// ────────────────────────────────────────────────
//

// Version returns the public Aether version string.
func Version() string {
	return fmt.Sprintf("Aether %s", version.AetherVersion)
}

// EffectiveConfig returns the final public configuration in effect for the client.
//
// Does not expose internal-only config fields.
func (c *Client) EffectiveConfig() Config {
	if c == nil || c.cfg == nil {
		return Config{}
	}

	return Config{
		// Networking
		UserAgent:          c.cfg.UserAgent,
		RequestTimeout:     c.cfg.RequestTimeout,
		MaxConcurrentHosts: c.cfg.MaxConcurrentHosts,
		MaxRequestsPerHost: c.cfg.MaxRequestsPerHost,

		// Logging
		EnableDebugLogging: c.cfg.EnableDebugLogging,

		// Caching
		EnableMemoryCache: c.cfg.EnableMemoryCache,
		EnableFileCache:   c.cfg.EnableFileCache,
		EnableRedisCache:  c.cfg.EnableRedisCache,

		CacheDirectory: c.cfg.CacheDirectory,
		RedisAddress:   c.cfg.RedisAddress,

		CacheTTL:        c.cfg.CacheTTL,
		MaxCacheEntries: c.cfg.MaxCacheEntries,
	}
}
