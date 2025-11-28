// cmd/test_crawl/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Nibir1/Aether/aether"
)

func main() {
	cli, err := aether.NewClient(
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Test : Crawl ===")

	startURL := "https://books.toscrape.com/"

	// Crawl visitor callback
	visitor := aether.CrawlVisitorFunc(func(ctx context.Context, p *aether.CrawledPage) error {
		fmt.Printf(
			"Visited: %s (depth=%d, status=%d, links=%d)\n",
			p.URL, p.Depth, p.StatusCode, len(p.Links),
		)
		return nil
	})

	opts := aether.CrawlOptions{
		MaxDepth:     1,
		MaxPages:     5,
		SameHostOnly: true,
		FetchDelay:   1 * time.Second, // throttle per-host
		Concurrency:  1,               // conservative for example.com
		Visitor:      visitor,
	}

	if err := cli.Crawl(ctx, startURL, opts); err != nil {
		log.Fatalf("Crawl error: %v", err)
	}
}
