// internal/crawl/perhost.go
//
// This file implements the per-host throttling logic for Aether's crawl
// subsystem. Even though Aether’s HTTP client enforces robots.txt compliance
// and per-host concurrency caps, ethical crawling also requires pacing:
//
//   • Never hammer a host with too many requests too quickly
//   • Allow callers to specify a minimum delay between hits to the same host
//
// This logic is independent of the HTTP client's concurrency-limiter. It
// enforces time-based spacing between visits per host, which can be layered
// on top of connection limits for stronger politeness guarantees.
//
// Implementation:
//   • A map of host → lastFetchTime
//   • Mutex-protected for concurrency
//   • A configurable FetchDelay (min time gap between requests)
//   • Waits (sleep) when needed before allowing the caller to continue

package crawl

import (
	"net/url"
	"sync"
	"time"
)

// PerHostThrottle enforces a minimum delay between requests to the same host.
//
// The crawler uses this to avoid overwhelming web servers even when multiple
// worker goroutines are active. This complements robots.txt compliance and the
// HTTP client's concurrency controls.
type PerHostThrottle struct {
	mu         sync.Mutex
	lastAccess map[string]time.Time
	minDelay   time.Duration
}

// NewPerHostThrottle constructs a new throttle enforcer.
//
// minDelay = 0 means "no throttling".
func NewPerHostThrottle(minDelay time.Duration) *PerHostThrottle {
	return &PerHostThrottle{
		lastAccess: make(map[string]time.Time),
		minDelay:   minDelay,
	}
}

// Wait respects the per-host delay before allowing another request to proceed.
//
// The caller should invoke Wait() *immediately before* performing a network
// fetch. This method blocks only the worker hitting this specific host.
// Workers hitting other hosts proceed unhindered.
func (p *PerHostThrottle) Wait(rawURL string) {
	if p.minDelay <= 0 {
		// Throttling disabled.
		return
	}

	host := extractHost(rawURL)
	if host == "" {
		// Unknown host → treat as no-throttle.
		return
	}

	p.mu.Lock()
	last, ok := p.lastAccess[host]
	now := time.Now()

	if ok {
		elapsed := now.Sub(last)
		if elapsed < p.minDelay {
			sleepFor := p.minDelay - elapsed
			p.mu.Unlock()
			time.Sleep(sleepFor)
			// After sleeping, update last-access timestamp.
			p.mu.Lock()
			p.lastAccess[host] = time.Now()
			p.mu.Unlock()
			return
		}
	}

	// No previous record or delay satisfied.
	p.lastAccess[host] = now
	p.mu.Unlock()
}

// extractHost parses the URL and returns the host portion.
//
// This helper is isolated here so it can evolve without affecting crawler.go.
func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Host
}
