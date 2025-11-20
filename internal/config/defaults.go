// internal/config/defaults.go
//
// This file centralizes default configuration constants used by Aether.
// Keeping them separate makes it easy to review and adjust the library's
// baseline behavior for networking, caching, and logging without touching
// any code that depends on Config.

package config

import "time"

const (
	// --- Networking defaults ---

	// defaultRequestTimeout is the baseline HTTP request timeout used
	// when callers do not specify a custom value.
	defaultRequestTimeout = 15 * time.Second

	// defaultMaxConcurrentHosts is the standard upper bound on the
	// number of distinct hosts Aether will contact concurrently.
	defaultMaxConcurrentHosts = 4

	// defaultMaxRequestsPerHost is the soft limit on in-flight requests
	// to a single host, used to avoid overwhelming any one site.
	defaultMaxRequestsPerHost = 4

	// --- Caching defaults (Stage 3) ---

	// defaultCacheTTL is the default time-to-live for all cache layers
	// (memory, file, redis) unless the user overrides it.
	defaultCacheTTL = 30 * time.Second

	// defaultMaxCacheEntries is the default capacity of the in-memory
	// LRU cache layer.
	defaultMaxCacheEntries = 128
)

// applyDefaults populates zero-valued fields in Config with library defaults.
//
// This helper is not currently used by Config.Default(), but it is kept
// as a central mechanism in case future functionality permits partially
// specified configurations.
//
// IMPORTANT:
// Config.Default() *already ensures* all fields have proper values.
// applyDefaults exists only for completeness and potential future use.
func applyDefaults(c *Config) {

	// Networking defaults
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = defaultRequestTimeout
	}
	if c.MaxConcurrentHosts <= 0 {
		c.MaxConcurrentHosts = defaultMaxConcurrentHosts
	}
	if c.MaxRequestsPerHost <= 0 {
		c.MaxRequestsPerHost = defaultMaxRequestsPerHost
	}

	// Caching defaults
	if c.CacheTTL <= 0 {
		c.CacheTTL = defaultCacheTTL
	}
	if c.MaxCacheEntries <= 0 {
		c.MaxCacheEntries = defaultMaxCacheEntries
	}

	// File cache directory defaults to empty → disabled
	// Redis address defaults to empty → disabled
	// Enable flags remain false unless explicitly set
}
