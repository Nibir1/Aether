// aether/aether.go
//
// Package aether provides the public entrypoint for the Aether library.
// Aether is a legal, robots.txt-compliant web retrieval toolkit that
// normalizes public, open data into JSON- and TOON-compatible models.
//
// As of Stage 13, Aether includes:
//   - unified composite caching (memory, file, redis)
//   - robots.txt-compliant HTTP fetcher
//   - HTML parsing + article extraction
//   - Detect/Meta subsystem
//   - SmartQuery router
//   - RSS/Atom subsystem
//   - OpenAPI integrations (Wikipedia, Wikidata, HN, GitHub, GovPress,
//     WhiteHouse, Weather via MET Norway)
//   - Public plugin system (SourcePlugins, TransformPlugins, DisplayPlugins)

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

	"github.com/Nibir1/Aether/plugins"
)

// DefaultUserAgent is the default HTTP User-Agent string that Aether uses
// when making outbound requests. It includes the repository for transparency.
const DefaultUserAgent = "AetherBot/1.0 (+https://github.com/Nibir1/Aether)"

// Client is the main public interface for using Aether.
//
// The Client owns:
//   - unified composite cache
//   - robots.txt-compliant HTTP fetcher
//   - internal OpenAPI client
//   - public plugin registry
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	fetcher *hclient.Client
	cache   icache.Cache
	openapi *iopenapi.Client

	plugins *plugins.Registry // internal plugin registry
}

// Config is the public, inspectable view of effective Aether configuration.
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

	// RobotsOverrideList exposes which hosts are configured to bypass
	// robots.txt checks. This is empty by default; any non-empty value
	// reflects explicit caller intent to override robots for those hosts.
	RobotsOverrideList []string
}

// Option is a functional option that modifies the internal configuration.
type Option func(*config.Config)

// NewClient constructs a new Aether Client with optional configuration.
//
// Pipeline:
//  1. Load default internal config
//  2. Apply user-specified Option functions
//  3. Ensure User-Agent is set
//  4. Initialize logger
//  5. Initialize unified composite cache
//  6. Initialize HTTP fetcher (robots.txt + caching)
//  7. Initialize internal OpenAPI client
//  8. Initialize plugin registry
func NewClient(opts ...Option) (*Client, error) {
	internalCfg := config.Default()

	for _, opt := range opts {
		if opt != nil {
			opt(internalCfg)
		}
	}

	// Default UA if caller did not provide one.
	if internalCfg.UserAgent == "" {
		internalCfg.UserAgent = DefaultUserAgent
	}

	logger := log.New(internalCfg.EnableDebugLogging)

	cli := &Client{
		cfg:     internalCfg,
		logger:  logger,
		plugins: plugins.NewRegistry(),
	}

	// unified composite cache
	cli.cache = icache.NewComposite(icache.Config{

		// Memory cache layer
		MemoryEnabled: internalCfg.EnableMemoryCache,
		MemoryTTL:     internalCfg.CacheTTL,
		MemoryMax:     internalCfg.MaxCacheEntries,

		// File cache layer
		FileEnabled:   internalCfg.EnableFileCache,
		FileTTL:       internalCfg.CacheTTL,
		FileDirectory: internalCfg.CacheDirectory,

		// Redis cache layer
		RedisEnabled: internalCfg.EnableRedisCache,
		RedisTTL:     internalCfg.CacheTTL,
		RedisAddress: internalCfg.RedisAddress,

		// Logging integration
		Logger: logger,
	})

	// robots.txt-compliant HTTP fetcher
	cli.fetcher = hclient.New(internalCfg, logger, cli.cache)

	// OpenAPI client
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

// WithRequestTimeout sets the HTTP timeout duration.
func WithRequestTimeout(d time.Duration) Option {
	return func(c *config.Config) {
		if d > 0 {
			c.RequestTimeout = d
		}
	}
}

// WithConcurrency sets concurrency caps for outbound HTTP requests.
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
func WithDebugLogging(enabled bool) Option {
	return func(c *config.Config) {
		c.EnableDebugLogging = enabled
	}
}

// WithRobotsOverride configures Aether to bypass robots.txt checks for
// the given hostnames.
//
// IMPORTANT:
//   - This does NOT disable robots globally.
//   - Robots are still respected for all other hosts.
//   - Hostnames are matched case-insensitively, without port.
//   - Responsibility for ignoring robots.txt lies with the caller.
//
// Example:
//
//	client, _ := aether.NewClient(
//	    aether.WithRobotsOverride("books.toscrape.com"),
//	)
func WithRobotsOverride(hosts ...string) Option {
	return func(c *config.Config) {
		if len(hosts) == 0 {
			return
		}
		c.RobotsOverrideList = append(c.RobotsOverrideList, hosts...)
	}
}

// WithRobotsAllowedHost registers a domain that is exempt from robots.txt.
// This is only honored when RobotsOverrideEnabled==true.
func WithRobotsAllowedHost(host string) Option {
	return func(c *config.Config) {
		if host == "" {
			return
		}
		c.RobotsAllowedHosts = append(c.RobotsAllowedHosts, host)
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

// EffectiveConfig returns the final public configuration in effect.
func (c *Client) EffectiveConfig() Config {
	if c == nil || c.cfg == nil {
		return Config{}
	}

	return Config{
		UserAgent:          c.cfg.UserAgent,
		RequestTimeout:     c.cfg.RequestTimeout,
		MaxConcurrentHosts: c.cfg.MaxConcurrentHosts,
		MaxRequestsPerHost: c.cfg.MaxRequestsPerHost,
		EnableDebugLogging: c.cfg.EnableDebugLogging,

		EnableMemoryCache: c.cfg.EnableMemoryCache,
		EnableFileCache:   c.cfg.EnableFileCache,
		EnableRedisCache:  c.cfg.EnableRedisCache,

		CacheDirectory:  c.cfg.CacheDirectory,
		RedisAddress:    c.cfg.RedisAddress,
		CacheTTL:        c.cfg.CacheTTL,
		MaxCacheEntries: c.cfg.MaxCacheEntries,

		RobotsOverrideList: append([]string(nil), c.cfg.RobotsOverrideList...),
	}
}
