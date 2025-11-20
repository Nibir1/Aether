// internal/httpclient/client.go
//
// Package httpclient implements Aether's internal HTTP client.
// It provides robots.txt-compliant HTTP GET with concurrency limits,
// unified caching (memory/file/redis), retry logic, and transparent
// User-Agent handling.
//
// Stage 3: Updated to use the full composite cache subsystem from
// internal/cache rather than the lightweight Stage 2 memory cache.

package httpclient

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Nibir1/Aether/internal/cache"
	"github.com/Nibir1/Aether/internal/config"
	"github.com/Nibir1/Aether/internal/errors"
	"github.com/Nibir1/Aether/internal/log"
)

// Error is Aether’s internal structured error type (re-exported).
type Error = errors.Error

// Client is Aether's internal HTTP client.
//
// NOTE:
// - Consumers do NOT use this directly.
// - The public aether.Client wraps this and provides Fetch().
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	http    *http.Client
	robots  *robotsCache
	limiter *hostLimiter
	cache   cache.Cache // NEW: unified memory/file/redis cache
}

// New constructs a new internal HTTP client.
//
// This client:
// - uses a single http.Client for pooling
// - respects timeouts
// - sets up robots.txt cache
// - sets concurrency limits
// - initializes the composite cache (memory/file/redis)
func New(cfg *config.Config, logger log.Logger, unified cache.Cache) *Client {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &Client{
		cfg:     cfg,
		logger:  logger,
		http:    httpClient,
		robots:  newRobotsCache(),
		limiter: newHostLimiter(cfg.MaxConcurrentHosts, cfg.MaxRequestsPerHost),
		cache:   unified, // unified composite cache
	}
}

// Fetch performs a robots.txt compliant HTTP GET with:
// - concurrency limiting
// - composite caching
// - retry logic
// - transparent User-Agent injection
//
// headers: optional additional request headers.
func (c *Client) Fetch(
	ctx context.Context,
	rawURL string,
	headers http.Header,
) (*Response, error) {

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.New(errors.KindHTTP, "invalid URL", err)
	}
	hostKey := parsed.Host

	// Global + per-host concurrency limiting.
	if err := c.limiter.Acquire(ctx, hostKey); err != nil {
		return nil, errors.New(errors.KindHTTP, "acquiring concurrency slot failed", err)
	}
	defer c.limiter.Release(hostKey)

	// Robots.txt check (fail-closed for errors, allow for fetch failures).
	allowed, err := c.robots.allowed(ctx, rawURL, c.cfg.UserAgent, c.http)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New(errors.KindRobots, "access disallowed by robots.txt", nil)
	}

	// ---- Composite Cache Check (memory → file → redis)
	cacheKey := "http:" + rawURL

	if c.cache != nil {
		if cached, ok := c.cache.Get(cacheKey); ok {
			c.logger.Debugf("cache hit (composite) for %s", rawURL)

			return &Response{
				URL:        rawURL,
				StatusCode: http.StatusOK,
				Header:     http.Header{"X-Aether-Cache": []string{"HIT"}},
				Body:       cached,
				FetchedAt:  time.Now(),
			}, nil
		}
	}

	// ---- Build headers for request
	reqHeaders := make(http.Header)
	for k, v := range headers {
		cp := make([]string, len(v))
		copy(cp, v)
		reqHeaders[k] = cp
	}
	reqHeaders.Set("User-Agent", c.cfg.UserAgent)
	if reqHeaders.Get("Accept") == "" {
		reqHeaders.Set("Accept", "*/*")
	}

	// ---- Retry Logic
	const maxRetries = 2
	backoff := 200 * time.Millisecond
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {

		// ctx cancellation short-circuit
		select {
		case <-ctx.Done():
			return nil, errors.New(errors.KindHTTP, "request canceled", ctx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, errors.New(errors.KindHTTP, "creating request failed", err)
		}
		req.Header = reqHeaders.Clone()

		resp, err := c.http.Do(req)
		if err != nil {
			if !isRetryableError(err) || attempt == maxRetries {
				return nil, errors.New(errors.KindHTTP, "request failed", err)
			}
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if attempt == maxRetries {
				return nil, errors.New(errors.KindHTTP, "reading response failed", readErr)
			}
			lastErr = readErr
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		out := &Response{
			URL:        rawURL,
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       body,
			FetchedAt:  time.Now(),
		}

		// ---- Store in unified cache (only cache 200 OK)
		if resp.StatusCode == http.StatusOK && c.cache != nil {
			c.cache.Set(cacheKey, body, c.cfg.CacheTTL)
		}

		return out, nil
	}

	if lastErr != nil {
		return nil, errors.New(errors.KindHTTP, "request failed after retries", lastErr)
	}
	return nil, errors.New(errors.KindHTTP, "request failed for unknown reasons", nil)
}

// isRetryableError reports whether the error is transient.
func isRetryableError(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout() || ne.Temporary()
	}
	return false
}
