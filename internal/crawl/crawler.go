// internal/crawl/crawler.go
//
// This file implements the main orchestrator for Aether's crawl subsystem.
// The crawler is responsible for:
//
//   • Managing the crawl frontier (queue of (URL, depth) pairs)
//   • Enforcing depth limits
//   • Respecting per-host throttling (politeness)
//   • Tracking visited URLs to avoid cycles
//   • Applying simple host/domain restrictions
//   • Fetching pages via Aether's internal HTTP client
//   • Extracting HTML links and scheduling child URLs
//   • Invoking a visitor callback with each crawled page
//
// Current implementation:
//
//   • Uses a single worker (sequential crawling). All building blocks
//     (frontier, throttle, visit map) are thread-safe and ready for a
//     future multi-worker version without changing the public API.
//
//   • Delegates robots.txt compliance, caching, retries, and rate limiting
//     to the internal httpclient.Client used by Aether.
//
// All crawling remains strictly legal and polite.

package crawl

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Nibir1/Aether/internal/httpclient"
)

// Page represents a single crawled page, as seen by the visitor callback.
type Page struct {
	URL        string
	Depth      int
	StatusCode int

	// Content is the raw response body interpreted as text. For non-text
	// content types, this will still be a string representation of bytes.
	Content string

	// Links contains the child URLs discovered on the page that were
	// accepted by host/domain/visited rules and enqueued for crawling.
	Links []string

	// Metadata holds additional simple metadata such as content type.
	Metadata map[string]string
}

// Visitor is invoked for each successfully fetched page.
type Visitor interface {
	VisitPage(ctx context.Context, page *Page) error
}

// VisitorFunc is a functional adapter to allow the use of
// ordinary functions as Visitors.
type VisitorFunc func(ctx context.Context, page *Page) error

// VisitPage calls f(ctx, page).
func (f VisitorFunc) VisitPage(ctx context.Context, page *Page) error {
	return f(ctx, page)
}

// Options configures the behavior of the crawler.
type Options struct {
	// MaxDepth is the maximum depth to crawl, starting at 0 for the root URL.
	// If MaxDepth < 0, depth is unlimited.
	MaxDepth int

	// MaxPages limits how many pages will be fetched. If MaxPages <= 0,
	// there is no explicit page limit and the crawl stops only when the
	// frontier becomes empty or the context is canceled.
	MaxPages int

	// SameHostOnly restricts all crawled URLs to the same host as the
	// starting URL.
	SameHostOnly bool

	// AllowedDomains, if non-empty, restricts crawling to these hostnames.
	// Hostnames are matched in their lowercase form.
	AllowedDomains []string

	// DisallowedDomains, if non-empty, blocks crawling for these hostnames.
	DisallowedDomains []string

	// FetchDelay is a soft politeness delay enforced between successive
	// requests to the same host. A value of zero disables per-host delay.
	FetchDelay time.Duration

	// Concurrency is reserved for a future multi-worker version of the
	// crawler. The current implementation uses a single worker, but keeps
	// this field for API stability.
	Concurrency int

	// Visitor is invoked for each fetched page. It must not be nil.
	Visitor Visitor
}

// Crawler is the internal crawl engine.
//
// It is constructed by the aether.Client and not exposed directly to
// end users. Public APIs will wrap this engine via aether/crawl.go.
type Crawler struct {
	fetcher *httpclient.Client
	opts    Options

	depthLimit DepthLimit
	frontier   *FrontierQueue
	visited    *VisitMap
	throttle   *PerHostThrottle

	startHost         string
	allowedDomains    map[string]struct{}
	disallowedDomains map[string]struct{}
}

// NewCrawler constructs a new Crawler using the provided internal HTTP
// client and options.
func NewCrawler(fetcher *httpclient.Client, opts Options) (*Crawler, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("crawl: nil HTTP fetcher")
	}
	if opts.Visitor == nil {
		return nil, fmt.Errorf("crawl: Visitor must not be nil")
	}

	c := &Crawler{
		fetcher:    fetcher,
		opts:       opts,
		depthLimit: NewDepthLimit(opts.MaxDepth),
		frontier:   NewFrontierQueue(),
		visited:    NewVisitMap(),
		throttle:   NewPerHostThrottle(opts.FetchDelay),
	}

	c.allowedDomains = make(map[string]struct{})
	for _, d := range opts.AllowedDomains {
		d = strings.TrimSpace(strings.ToLower(d))
		if d != "" {
			c.allowedDomains[d] = struct{}{}
		}
	}

	c.disallowedDomains = make(map[string]struct{})
	for _, d := range opts.DisallowedDomains {
		d = strings.TrimSpace(strings.ToLower(d))
		if d != "" {
			c.disallowedDomains[d] = struct{}{}
		}
	}

	return c, nil
}

// Run executes the crawl starting from startURL.
//
// The crawl stops when:
//   - the frontier is empty, or
//   - MaxPages (if > 0) is reached, or
//   - the context is canceled, or
//   - a fatal error is returned by the Visitor or fetcher.
//
// The current implementation is single-threaded (one worker), but all
// underlying components are safe for future multi-worker expansion.
func (c *Crawler) Run(ctx context.Context, startURL string) error {
	norm, host, err := c.normalizeStartURL(startURL)
	if err != nil {
		return err
	}
	c.startHost = host

	if !c.hostAllowed(host) {
		return fmt.Errorf("crawl: start host %q is not allowed", host)
	}

	// Seed frontier with the root URL at depth 0.
	c.frontier.Enqueue(FrontierItem{
		URL:   norm,
		Depth: 0,
	})
	c.visited.MarkVisited(norm)

	pagesFetched := 0

	for {
		// Respect context cancellation.
		if ctx.Err() != nil {
			return ctx.Err()
		}

		item, ok := c.frontier.Dequeue()
		if !ok {
			// No more URLs to visit → crawl complete.
			return nil
		}

		if !c.depthLimit.Allowed(item.Depth) {
			continue
		}

		if !c.hostAllowed(extractHost(item.URL)) {
			continue
		}

		if c.opts.MaxPages > 0 && pagesFetched >= c.opts.MaxPages {
			return nil
		}

		// Politeness delay per host.
		c.throttle.Wait(item.URL)

		resp, err := c.fetcher.Fetch(ctx, item.URL, nil)
		if err != nil {
			return err
		}

		pagesFetched++

		contentType := ""
		if resp.Header != nil {
			contentType = resp.Header.Get("Content-Type")
		}
		body := string(resp.Body)

		page := &Page{
			URL:        item.URL,
			Depth:      item.Depth,
			StatusCode: resp.StatusCode,
			Content:    body,
			Metadata: map[string]string{
				"content_type": contentType,
			},
		}

		// Extract child links only for HTML content.
		if strings.Contains(strings.ToLower(contentType), "html") {
			baseURL, _ := url.Parse(item.URL)
			links := extractLinks(baseURL, body)
			page.Links = c.filterAndEnqueueChildren(links, item.Depth)
		}

		if err := c.opts.Visitor.VisitPage(ctx, page); err != nil {
			return err
		}
	}
}

// normalizeStartURL normalizes the starting URL into an absolute, canonical form
// and returns (normalizedURL, host, error).
func (c *Crawler) normalizeStartURL(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("crawl: empty start URL")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("crawl: invalid start URL: %w", err)
	}
	if !u.IsAbs() {
		return "", "", fmt.Errorf("crawl: start URL must be absolute (got %q)", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", fmt.Errorf("crawl: unsupported scheme %q", u.Scheme)
	}

	u.Fragment = ""
	u.Host = strings.ToLower(u.Host)

	return u.String(), u.Host, nil
}

// hostAllowed checks whether a host passes SameHostOnly, allowed, and
// disallowed domain rules.
func (c *Crawler) hostAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}

	if c.opts.SameHostOnly && c.startHost != "" && host != c.startHost {
		return false
	}

	if _, blocked := c.disallowedDomains[host]; blocked {
		return false
	}

	if len(c.allowedDomains) > 0 {
		if _, ok := c.allowedDomains[host]; !ok {
			return false
		}
	}

	return true
}

// filterAndEnqueueChildren normalizes discovered links, applies host/domain
// rules, depth rules, and visited-set checks, enqueues valid children into
// the frontier, and returns the list of accepted child URLs.
func (c *Crawler) filterAndEnqueueChildren(links []string, parentDepth int) []string {
	if len(links) == 0 {
		return nil
	}

	nextDepth := c.depthLimit.Next(parentDepth)
	if !c.depthLimit.Allowed(nextDepth) {
		return nil
	}

	accepted := make([]string, 0, len(links))

	for _, raw := range links {
		norm, host := normalizeURLForChild(raw)
		if norm == "" || host == "" {
			continue
		}
		if !c.hostAllowed(host) {
			continue
		}
		if !c.visited.MarkVisited(norm) {
			continue
		}

		c.frontier.Enqueue(FrontierItem{
			URL:   norm,
			Depth: nextDepth,
		})
		accepted = append(accepted, norm)
	}

	if len(accepted) == 0 {
		return nil
	}
	return accepted
}

// normalizeURLForChild normalizes a child URL string into an absolute form
// when it is already absolute. Relative resolution is handled by extractLinks.
// Here we simply parse and canonicalize scheme, host, and fragment.
func normalizeURLForChild(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", ""
	}
	if !u.IsAbs() {
		// Relative URLs should already have been resolved in extractLinks.
		return "", ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ""
	}
	u.Fragment = ""
	u.Host = strings.ToLower(u.Host)

	return u.String(), u.Host
}

// extractLinks performs a simple, conservative extraction of links from
// an HTML document by scanning for href="..." or href='...' attributes.
//
// It resolves relative links against the provided base URL and returns
// a de-duplicated slice of absolute URLs.
//
// This is intentionally minimal and does not attempt to be a perfect
// HTML parser; Aether's more advanced HTML parsing pipeline is used
// elsewhere when deep extraction is required.
func extractLinks(base *url.URL, htmlBody string) []string {
	if base == nil {
		return nil
	}

	var links []string
	s := htmlBody
	i := 0

	for {
		idx := strings.Index(s[i:], "href=")
		if idx < 0 {
			break
		}
		idx += i
		if idx+6 > len(s) {
			break
		}

		quote := s[idx+5]
		if quote != '"' && quote != '\'' {
			i = idx + 5
			continue
		}

		start := idx + 6
		end := strings.IndexByte(s[start:], quote)
		if end < 0 {
			break
		}
		end = start + end

		href := strings.TrimSpace(s[start:end])
		if href != "" {
			if abs, _ := resolveRelativeURL(base, href); abs != "" {
				links = append(links, abs)
			}
		}

		i = end + 1
	}

	if len(links) == 0 {
		return nil
	}

	// De-duplicate.
	seen := make(map[string]struct{}, len(links))
	out := make([]string, 0, len(links))
	for _, l := range links {
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}

	return out
}

// resolveRelativeURL resolves href against base and returns an absolute,
// canonicalized URL and its host.
func resolveRelativeURL(base *url.URL, href string) (string, string) {
	u, err := url.Parse(href)
	if err != nil {
		return "", ""
	}
	if !u.IsAbs() {
		u = base.ResolveReference(u)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ""
	}
	u.Fragment = ""
	u.Host = strings.ToLower(u.Host)

	return u.String(), u.Host
}
