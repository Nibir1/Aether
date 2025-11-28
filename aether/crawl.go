// aether/crawl.go
//
// Public crawl API for Aether.
//
// This file provides a stable, legal, user-facing interface for crawling
// public websites in a robots.txt-compliant, polite manner.

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

// CrawledPage is the public representation of a fetched page.
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

// CrawlVisitorFunc adapts functions to CrawlVisitor.
type CrawlVisitorFunc func(ctx context.Context, page *CrawledPage) error

func (f CrawlVisitorFunc) VisitCrawledPage(ctx context.Context, page *CrawledPage) error {
	return f(ctx, page)
}

// CrawlOptions configures Aether's crawl behavior.
type CrawlOptions struct {
	MaxDepth          int
	MaxPages          int
	SameHostOnly      bool
	AllowedDomains    []string
	DisallowedDomains []string
	FetchDelay        time.Duration
	Concurrency       int
	Visitor           CrawlVisitor
}

//
// ─────────────────────────────────────────────
//            CLIENT PUBLIC METHOD
// ─────────────────────────────────────────────
//

// Crawl launches a polite, robots.txt-compliant crawl.
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

	// Normalize concurrency
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}

	// Convert public → internal options
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

	// Create crawler engine (robots-compliant)
	engine, err := icrawl.NewCrawler(c.fetcher, intOpts)
	if err != nil {
		return err
	}

	// Execute crawl
	return engine.Run(ctx, startURL)
}

//
// ─────────────────────────────────────────────
//            VISITOR ADAPTER LAYER
// ─────────────────────────────────────────────
//

// crawlVisitorAdapter converts internal.Page → public CrawledPage.
type crawlVisitorAdapter struct {
	pub CrawlVisitor
}

// This MUST match internal/crawl.Visitor's method name & signature exactly.
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
