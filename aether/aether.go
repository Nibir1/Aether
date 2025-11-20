// aether/aether.go
//
// Package aether provides the public entrypoint for the Aether library.
// Aether is a legal, robots.txt-compliant web retrieval toolkit that
// normalizes public, open data into a JSON-compatible model and can
// serialize it to both JSON and TOON formats for LLM usage.
//
// Stage 1: This file defines the root Client, configuration options,
// and a version helper.
// Stage 2: The Client now owns an internal HTTP fetcher that implements
// robots.txt-compliant HTTP GET with concurrency limits and caching.
package aether

import (
	"fmt"
	"time"

	"github.com/Nibir1/Aether/internal/config"
	hclient "github.com/Nibir1/Aether/internal/httpclient"
	"github.com/Nibir1/Aether/internal/log"
	"github.com/Nibir1/Aether/internal/version"
)

// DefaultUserAgent is the default HTTP User-Agent string that Aether uses
// when making outbound HTTP requests.
//
// The string clearly identifies the library and repository, which is an
// important part of responsible, transparent web access.
const DefaultUserAgent = "AetherBot/1.0 (+https://github.com/Nibir1/Aether)"

// Client is the main public handle for using Aether.
//
// All high-level capabilities such as Search, Lookup, Batch and Crawl
// will be implemented as methods on Client in later stages. As of Stage 2,
// Client also encapsulates an internal HTTP fetcher.
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	fetcher *hclient.Client
}

// Config is the public view of Aether configuration.
//
// This type mirrors the internal config.Config fields that we want to
// expose. Keeping it separate from the internal type allows us to evolve
// implementation details without breaking user code.
type Config struct {
	UserAgent          string
	RequestTimeout     time.Duration
	MaxConcurrentHosts int
	MaxRequestsPerHost int
	EnableDebugLogging bool
}

// Option is a functional option used to customize the Client.
//
// This pattern keeps the constructor flexible and backward compatible
// as Aether grows new configuration settings.
type Option func(*config.Config)

// NewClient constructs a new Aether Client with optional configuration.
//
// It starts from the internal default configuration, applies all
// provided Option functions, ensures a reasonable User-Agent string is
// set, initializes a logger and constructs an internal HTTP fetcher.
//
// At Stage 2, the Client can perform robots.txt-compliant HTTP GET
// requests via the Fetch method.
func NewClient(opts ...Option) (*Client, error) {
	internalCfg := config.Default()

	for _, opt := range opts {
		if opt != nil {
			opt(internalCfg)
		}
	}

	// Ensure a transparent, well-formed User-Agent is always present.
	if internalCfg.UserAgent == "" {
		internalCfg.UserAgent = DefaultUserAgent
	}

	logger := log.New(internalCfg.EnableDebugLogging)
	fetcher := hclient.New(internalCfg, logger)

	return &Client{
		cfg:     internalCfg,
		logger:  logger,
		fetcher: fetcher,
	}, nil
}

// WithUserAgent overrides the default User-Agent string.
//
// This allows applications to identify themselves more specifically,
// while still respecting Aether's legal/ethical responsibility to be
// transparent about the client identity when accessing public sites.
func WithUserAgent(ua string) Option {
	return func(c *config.Config) {
		if ua != "" {
			c.UserAgent = ua
		}
	}
}

// WithRequestTimeout sets the HTTP request timeout used by Aether's
// network operations.
//
// Very short timeouts may cause frequent failures; very long timeouts
// may delay recovery from issues. Sensible values are typically in the
// 5–60 second range, depending on the application.
func WithRequestTimeout(d time.Duration) Option {
	return func(c *config.Config) {
		if d > 0 {
			c.RequestTimeout = d
		}
	}
}

// WithConcurrency configures basic concurrency limits for Aether's
// outbound HTTP operations.
//
// maxHosts controls how many distinct hosts Aether may contact
// concurrently, and maxPerHost sets a soft limit on how many requests
// can be in flight for a single host. These limits help ensure polite,
// robots.txt-compliant traffic patterns.
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

// WithDebugLogging enables or disables verbose internal logging.
//
// Debug logs are extremely useful while integrating or developing with
// Aether, but may be too noisy for production environments.
func WithDebugLogging(enabled bool) Option {
	return func(c *config.Config) {
		c.EnableDebugLogging = enabled
	}
}

// Version returns a human-readable Aether version string.
//
// The underlying semantic version is maintained in an internal package,
// which allows the public API to present it in a stable way.
func Version() string {
	return fmt.Sprintf("Aether %s", version.AetherVersion)
}

// EffectiveConfig returns a copy of the current configuration as a
// public Config value.
//
// Callers can inspect this to understand which defaults and options
// are in effect for a given Client instance.
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
	}
}
