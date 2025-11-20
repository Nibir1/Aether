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
// future stages will extend this struct with caching, robots.txt, and
// normalization options as needed.
type Config struct {
	// UserAgent is the HTTP User-Agent string Aether uses for outbound
	// requests. It should clearly identify Aether and, ideally, the
	// integrating application.
	UserAgent string

	// RequestTimeout is the maximum amount of time Aether will wait for
	// a single HTTP request to complete.
	RequestTimeout time.Duration

	// MaxConcurrentHosts is an upper bound on how many different hosts
	// Aether will talk to concurrently. This contributes to polite,
	// non-aggressive network behavior.
	MaxConcurrentHosts int

	// MaxRequestsPerHost is a soft limit for in-flight requests to the
	// same host, used to avoid overwhelming any single domain.
	MaxRequestsPerHost int

	// EnableDebugLogging controls whether verbose internal logging is
	// enabled. Debug logging is useful during development and debugging,
	// but may be too noisy for production usage.
	EnableDebugLogging bool

	// CacheTTL controls how long successful GET responses may be kept
	// in the in-memory HTTP cache before they are considered stale.
	CacheTTL time.Duration

	// MaxCacheEntries bounds the number of entries stored in the
	// in-memory HTTP cache. When this limit is exceeded, older entries
	// are evicted in a simple best-effort manner.
	MaxCacheEntries int
}

// Default constructs a Config with safe, conservative defaults.
//
// These defaults are chosen to be reasonable for a wide range of
// applications. Callers can adjust them through the public functional
// options defined in the aether package.
// Default constructs a Config with safe, conservative defaults.
func Default() *Config {
	return &Config{
		UserAgent:          "", // filled by aether.NewClient if empty
		RequestTimeout:     defaultRequestTimeout,
		MaxConcurrentHosts: defaultMaxConcurrentHosts,
		MaxRequestsPerHost: defaultMaxRequestsPerHost,
		EnableDebugLogging: false,
		CacheTTL:           defaultCacheTTL,
		MaxCacheEntries:    defaultMaxCacheEntries,
	}
}
