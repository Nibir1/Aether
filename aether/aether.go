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

// DefaultUserAgent is the default HTTP User-Agent string.
const DefaultUserAgent = "AetherBot/1.0 (+https://github.com/Nibir1/Aether)"

// Client is Aether’s main entrypoint.
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	fetcher *hclient.Client
	cache   icache.Cache
	openapi *iopenapi.Client
	plugins *plugins.Registry
}

// Public effective config returned to the user.
type Config struct {
	UserAgent          string
	RequestTimeout     time.Duration
	MaxConcurrentHosts int
	MaxRequestsPerHost int
	EnableDebugLogging bool

	EnableMemoryCache bool
	EnableFileCache   bool
	EnableRedisCache  bool

	CacheDirectory string
	RedisAddress   string

	CacheTTL        time.Duration
	MaxCacheEntries int
}

type Option func(*config.Config)

//
// ─────────────────────────────────────────────
//                NEW CLIENT
// ─────────────────────────────────────────────
//

func NewClient(opts ...Option) (*Client, error) {
	// 1) load default internal config
	internalCfg := config.Default()

	// 2) apply all Option modifiers
	for _, opt := range opts {
		if opt != nil {
			opt(internalCfg)
		}
	}

	// 3) ensure default user-agent
	if internalCfg.UserAgent == "" {
		internalCfg.UserAgent = DefaultUserAgent
	}

	// 4) initialize logger (returns ONLY 1 value)
	logger := log.New(internalCfg.EnableDebugLogging)

	cli := &Client{
		cfg:     internalCfg,
		logger:  logger,
		plugins: plugins.NewRegistry(),
	}

	// 5) unified composite cache (returns ONLY 1 value)
	cli.cache = icache.NewComposite(icache.Config{
		MemoryEnabled: internalCfg.EnableMemoryCache,
		MemoryTTL:     internalCfg.CacheTTL,
		MemoryMax:     internalCfg.MaxCacheEntries,

		FileEnabled:   internalCfg.EnableFileCache,
		FileTTL:       internalCfg.CacheTTL,
		FileDirectory: internalCfg.CacheDirectory,

		RedisEnabled: internalCfg.EnableRedisCache,
		RedisTTL:     internalCfg.CacheTTL,
		RedisAddress: internalCfg.RedisAddress,

		Logger: logger,
	})

	// 6) robots.txt-compliant HTTP fetcher (returns ONLY 1 value)
	cli.fetcher = hclient.New(internalCfg, logger, cli.cache)

	// 7) OpenAPI client (returns ONLY 1 value)
	cli.openapi = iopenapi.New(internalCfg, logger, cli.fetcher)

	return cli, nil
}

//
// ─────────────────────────────────────────────
//            PUBLIC CONFIGURATION OPTIONS
// ─────────────────────────────────────────────
//

func WithUserAgent(ua string) Option {
	return func(c *config.Config) {
		if ua != "" {
			c.UserAgent = ua
		}
	}
}

func WithRequestTimeout(d time.Duration) Option {
	return func(c *config.Config) {
		if d > 0 {
			c.RequestTimeout = d
		}
	}
}

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

func WithDebugLogging(enabled bool) Option {
	return func(c *config.Config) {
		c.EnableDebugLogging = enabled
	}
}

//
// ─────────────────────────────────────────────
//                 PUBLIC HELPERS
// ─────────────────────────────────────────────
//

func Version() string {
	return fmt.Sprintf("Aether %s", version.AetherVersion)
}

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
		CacheDirectory:    c.cfg.CacheDirectory,
		RedisAddress:      c.cfg.RedisAddress,
		CacheTTL:          c.cfg.CacheTTL,
		MaxCacheEntries:   c.cfg.MaxCacheEntries,
	}
}
