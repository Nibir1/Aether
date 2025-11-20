// aether/crawl.go
//
// Public crawl API for Aether.
//
// This file provides a stable, legal, user-facing interface for crawling
// public websites in a robots.txt-compliant, polite manner. The internal
// crawl engine lives under internal/crawl and is not exposed directly.
//
// Aether's crawler is intentionally conservative:
//   - obeys robots.txt (via internal httpclient)
//   - respects per-host throttling delays
//   - enforces depth limits
//   - enforces same-host and domain restrictions
//   - ensures no duplicate URLs via visited-set tracking
//
// This public API is what users call via:
//
//     client := aether.NewClient(...)
//     err := client.Crawl(ctx, "https://example.com", opts)
//
// Future versions may add multi-worker concurrency, richer link extraction,
// or plugin-based crawl transform stages — without breaking this API.

package aether

import (
	"context"
	"fmt"
	"strings"
	"time"

	icrawl "github.com/Nibir1/Aether/internal/crawl"
)

//
// ─────────────────────────────────────────────
//              PUBLIC TYPES
// ─────────────────────────────────────────────
//

// CrawledPage is the public representation of a fetched page during crawling.
// This is a thin wrapper around internal/crawl.Page.
type CrawledPage struct {
	URL        string
	Depth      int
	StatusCode int
	Content    string
	Links      []string
	Metadata   map[string]string
}

// CrawlVisitor defines the callback interface for receiving crawled pages.
type CrawlVisitor interface {
	VisitCrawledPage(ctx context.Context, page *CrawledPage) error
}

// CrawlVisitorFunc is a functional adapter for using ordinary functions.
type CrawlVisitorFunc func(ctx context.Context, page *CrawledPage) error

// VisitCrawledPage calls f(ctx, page).
func (f CrawlVisitorFunc) VisitCrawledPage(ctx context.Context, page *CrawledPage) error {
	return f(ctx, page)
}

// CrawlOptions configures the behavior of Aether's public crawl API.
type CrawlOptions struct {
	// Maximum link depth to traverse (0 = only the start page).
	// If MaxDepth < 0, depth is unlimited.
	MaxDepth int

	// Maximum number of pages to fetch. If <= 0, unlimited.
	MaxPages int

	// Restrict crawling to the same host as the start URL.
	SameHostOnly bool

	// Additional domain_allow list; empty means "no restrictions".
	AllowedDomains []string

	// Optional domain blocklist.
	DisallowedDomains []string

	// Minimum delay between requests to the same host.
	FetchDelay time.Duration

	// Future-proof concurrency parameter.
	Concurrency int

	// Callback invoked for each visited page.
	Visitor CrawlVisitor
}

//
// ─────────────────────────────────────────────
//            CLIENT PUBLIC METHOD
// ─────────────────────────────────────────────
//

// Crawl executes a legal, robots.txt-compliant crawl starting at startURL,
// using the provided options.
//
// The crawl stops when:
//   - frontier is empty, or
//   - MaxPages is reached, or
//   - context is canceled, or
//   - Visitor returns an error.
func (c *Client) Crawl(ctx context.Context, startURL string, opts CrawlOptions) error {
	if c == nil {
		return fmt.Errorf("aether: nil client in Crawl")
	}
	if strings.TrimSpace(startURL) == "" {
		return fmt.Errorf("aether: empty startURL in Crawl")
	}
	if opts.Visitor == nil {
		return fmt.Errorf("aether: CrawlOptions.Visitor must not be nil")
	}

	// Convert public options → internal crawl.Options.
	intOpts := icrawl.Options{
		MaxDepth:          opts.MaxDepth,
		MaxPages:          opts.MaxPages,
		SameHostOnly:      opts.SameHostOnly,
		AllowedDomains:    opts.AllowedDomains,
		DisallowedDomains: opts.DisallowedDomains,
		FetchDelay:        opts.FetchDelay,
		Concurrency:       opts.Concurrency,
		Visitor: &crawlVisitorAdapter{
			pub: opts.Visitor,
		},
	}

	engine, err := icrawl.NewCrawler(c.fetcher, intOpts)
	if err != nil {
		return err
	}

	return engine.Run(ctx, startURL)
}

//
// ─────────────────────────────────────────────
//           VISITOR ADAPTER LAYER
// ─────────────────────────────────────────────
//
// The internal crawler expects a Visitor interface with a slightly
// different method signature. This adapter converts internal.Page →
// public CrawledPage and forwards it to the user's CrawlVisitor.
//

type crawlVisitorAdapter struct {
	pub CrawlVisitor
}

func (a *crawlVisitorAdapter) VisitPage(ctx context.Context, p *icrawl.Page) error {
	if a.pub == nil || p == nil {
		return nil
	}

	pub := &CrawledPage{
		URL:        p.URL,
		Depth:      p.Depth,
		StatusCode: p.StatusCode,
		Content:    p.Content,
		Links:      p.Links,
		Metadata:   p.Metadata,
	}

	return a.pub.VisitCrawledPage(ctx, pub)
}
