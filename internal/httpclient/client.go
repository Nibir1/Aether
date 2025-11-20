// internal/httpclient/client.go
//
// Package httpclient implements Aether's internal HTTP client.
// It provides robots.txt-compliant HTTP GET with concurrency limits,
// basic in-memory caching and retry logic.
package httpclient

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Nibir1/Aether/internal/config"
	"github.com/Nibir1/Aether/internal/errors"
	"github.com/Nibir1/Aether/internal/log"
)

// Error is a convenient alias for the structured error type used by
// the HTTP client. It matches Aether's public Error type.
type Error = errors.Error

// Client is Aether's internal HTTP client.
//
// It should not be used directly by consumers of the aether package;
// instead, they call Client.Fetch at the aether level.
type Client struct {
	cfg     *config.Config
	logger  log.Logger
	http    *http.Client
	robots  *robotsCache
	limiter *hostLimiter
	cache   *memoryCache
}

// New constructs a new HTTP client with the provided configuration
// and logger. It reuses a single http.Client to benefit from connection
// pooling.
func New(cfg *config.Config, logger log.Logger) *Client {
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
		cache:   newMemoryCache(cfg.CacheTTL, cfg.MaxCacheEntries),
	}
}

// Fetch performs a robots.txt-compliant HTTP GET with retries,
// concurrency limiting and basic caching.
//
// headers may contain additional headers to send. The User-Agent header
// will always be set to the configured Aether User-Agent, overriding any
// User-Agent value in headers.
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

	// Concurrency limiting.
	if err := c.limiter.Acquire(ctx, hostKey); err != nil {
		return nil, errors.New(errors.KindHTTP, "acquiring concurrency slot failed", err)
	}
	defer c.limiter.Release(hostKey)

	// robots.txt check.
	allowed, err := c.robots.allowed(ctx, rawURL, c.cfg.UserAgent, c.http)
	if err != nil {
		// conservative: treat failure as an HTTP error.
		return nil, err
	}
	if !allowed {
		return nil, errors.New(errors.KindRobots, "access disallowed by robots.txt", nil)
	}

	// Cache check.
	if resp := c.cache.Get(rawURL); resp != nil {
		c.logger.Debugf("cache hit for %s", rawURL)
		return resp, nil
	}

	// Build base request (without context; context is applied per attempt).
	reqHeaders := make(http.Header)
	for k, v := range headers {
		cp := make([]string, len(v))
		copy(cp, v)
		reqHeaders[k] = cp
	}
	// Ensure User-Agent is set to Aether's configured value.
	reqHeaders.Set("User-Agent", c.cfg.UserAgent)
	if reqHeaders.Get("Accept") == "" {
		reqHeaders.Set("Accept", "*/*")
	}

	const maxRetries = 2
	backoff := 200 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
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

		// Cache successful 200 responses.
		if resp.StatusCode == http.StatusOK {
			c.cache.Set(rawURL, out)
		}

		return out, nil
	}

	// Fallback in case the loop exits without returning.
	if lastErr != nil {
		return nil, errors.New(errors.KindHTTP, "request failed after retries", lastErr)
	}
	return nil, errors.New(errors.KindHTTP, "request failed for unknown reasons", nil)
}

// isRetryableError reports whether the error is likely transient.
func isRetryableError(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout() || ne.Temporary()
	}
	return false
}
