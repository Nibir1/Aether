// internal/config/config.go
//
// Package config defines internal configuration structures for Aether.
// This package is internal to preserve flexibility: we may add, change,
// or remove fields without breaking the public API as long as the
// externally visible behavior of aether.Client remains consistent.

package config

import "time"

// Config holds core configuration values used across Aether.
//
// Many of these values control networking behavior and logging verbosity;
// Stage 3 extends this struct with caching configuration.
type Config struct {

	// HTTP settings
	UserAgent          string
	RequestTimeout     time.Duration
	MaxConcurrentHosts int
	MaxRequestsPerHost int

	// Logging
	EnableDebugLogging bool

	// --- Caching settings (Stage 3) ---

	// CacheTTL is the default time-to-live for all cache layers.
	CacheTTL time.Duration

	// MaxCacheEntries affects memory LRU capacity.
	MaxCacheEntries int

	// Memory cache enabled flag
	EnableMemoryCache bool

	// File cache enabled flag
	EnableFileCache bool

	// Directory where file cache entries are stored
	CacheDirectory string

	// Redis cache enabled flag
	EnableRedisCache bool

	// Redis server address, e.g. "localhost:6379"
	RedisAddress string
}

// Default constructs a Config with safe, conservative defaults.
//
// These defaults are chosen to support polite network behavior and
// predictable caching without requiring configuration.
func Default() *Config {
	return &Config{
		UserAgent:          "",
		RequestTimeout:     defaultRequestTimeout,
		MaxConcurrentHosts: defaultMaxConcurrentHosts,
		MaxRequestsPerHost: defaultMaxRequestsPerHost,
		EnableDebugLogging: false,

		// cache defaults
		CacheTTL:        defaultCacheTTL,
		MaxCacheEntries: defaultMaxCacheEntries,

		EnableMemoryCache: false,
		EnableFileCache:   false,
		EnableRedisCache:  false,

		CacheDirectory: "",
		RedisAddress:   "",
	}
}
