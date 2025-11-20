// aether/aether.go
//
// Package aether provides the public entrypoint for the Aether library.
// Aether is a legal, robots.txt-compliant web retrieval toolkit that
// normalizes public, open data into JSON- and TOON-compatible models.
//
// Stage 3:
// The Client now owns a unified composite cache (memory, file, redis),
// and an internal HTTP fetcher that uses that cache together with a
// robots.txt-compliant request pipeline.

package aether

import (
	"fmt"
	"time"

	icache "github.com/Nibir1/Aether/internal/cache"
	"github.com/Nibir1/Aether/internal/config"
	hclient "github.com/Nibir1/Aether/internal/httpclient"
	"github.com/Nibir1/Aether/internal/log"
	"github.com/Nibir1/Aether/internal/version"
)

// DefaultUserAgent is the default HTTP User-Agent string that Aether uses
// for outbound requests. It identifies the library responsibly.
const DefaultUserAgent = "AetherBot/1.0 (+https://github.com/Nibir1/Aether)"

// Client is the main public interface for using Aether.
//
// Stage 3:
// - owns unified cache (memory + file + redis)
// - owns internal HTTP fetcher
// - will own parsers, search pipeline, TOON/JSON normalizers in later stages
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	fetcher *hclient.Client
	cache   icache.Cache
}

// Config is the public, inspectable view of Aether configuration.
//
// This mirrors internal config.Config, but exposes only fields intended for
// public visibility and does not reveal internal implementation details.
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
type Option func(*config.Config)

// NewClient constructs a new Aether Client with optional configuration.
//
// Steps:
// 1. Load internal defaults.
// 2. Apply functional options.
// 3. Ensure a User-Agent is set.
// 4. Initialize logger.
// 5. Initialize composite cache.
// 6. Initialize internal HTTP fetcher (robots + cache + concurrency).
func NewClient(opts ...Option) (*Client, error) {
	internalCfg := config.Default()

	for _, opt := range opts {
		if opt != nil {
			opt(internalCfg)
		}
	}

	// Default User-Agent if none provided.
	if internalCfg.UserAgent == "" {
		internalCfg.UserAgent = DefaultUserAgent
	}

	logger := log.New(internalCfg.EnableDebugLogging)

	cli := &Client{
		cfg:    internalCfg,
		logger: logger,
	}

	// Build unified composite cache (memory + file + redis).
	cli.initCache()

	// Build HTTP fetcher with unified cache.
	cli.fetcher = hclient.New(internalCfg, logger, cli.cache)

	return cli, nil
}

// WithUserAgent overrides the HTTP User-Agent Aether will send.
func WithUserAgent(ua string) Option {
	return func(c *config.Config) {
		if ua != "" {
			c.UserAgent = ua
		}
	}
}

// WithRequestTimeout sets the timeout for HTTP GET operations.
func WithRequestTimeout(d time.Duration) Option {
	return func(c *config.Config) {
		if d > 0 {
			c.RequestTimeout = d
		}
	}
}

// WithConcurrency configures concurrency limits for outbound network I/O.
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

// WithDebugLogging enables verbose internal logging.
func WithDebugLogging(enabled bool) Option {
	return func(c *config.Config) {
		c.EnableDebugLogging = enabled
	}
}

// Version returns the Aether version as a string.
func Version() string {
	return fmt.Sprintf("Aether %s", version.AetherVersion)
}

// EffectiveConfig returns the final public configuration in effect.
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
