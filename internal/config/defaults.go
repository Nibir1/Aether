// internal/config/defaults.go
//
// This file centralizes default configuration constants used by Aether.
// Keeping them separate makes it easy to review and adjust the library's
// baseline behavior for networking and logging without touching code
// that depends on Config.
package config

import "time"

const (
	// defaultRequestTimeout is the baseline HTTP request timeout used
	// when callers do not specify a custom value.
	defaultRequestTimeout = 15 * time.Second

	// defaultMaxConcurrentHosts is the standard upper bound on the
	// number of distinct hosts Aether will contact concurrently.
	defaultMaxConcurrentHosts = 4

	// defaultMaxRequestsPerHost is the default limit on in-flight
	// requests to a single host, used to avoid overwhelming any one site.
	defaultMaxRequestsPerHost = 4

	// defaultCacheTTL is the default time-to-live for cached HTTP
	// responses in the in-memory cache.
	defaultCacheTTL = 30 * time.Second

	// defaultMaxCacheEntries is the default upper bound on the number
	// of responses kept in the in-memory HTTP cache.
	defaultMaxCacheEntries = 128
)

// applyDefaults populates zero-valued fields in Config with the library's
// standard defaults. It may be used by future helpers if we allow
// partially-specified configurations.
//
// At Stage 1, Config.Default already sets all fields, but this helper
// illustrates how we would centralize default logic.
func applyDefaults(c *Config) {
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = defaultRequestTimeout
	}
	if c.MaxConcurrentHosts <= 0 {
		c.MaxConcurrentHosts = defaultMaxConcurrentHosts
	}
	if c.MaxRequestsPerHost <= 0 {
		c.MaxRequestsPerHost = defaultMaxRequestsPerHost
	}
}
