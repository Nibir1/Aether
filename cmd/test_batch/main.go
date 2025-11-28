// cmd/test_batch/main.go
//
// This file implements a small standalone batch-fetch test that works
// with the current Aether public API. Aether does NOT (yet) provide a
// built-in Batch() function, so we implement a small wrapper here using
// concurrency + cli.Fetch().
//
// This test demonstrates:
//   • Robots.txt-compliant fetch
//   • Per-request headers
//   • Concurrency control
//   • Graceful error handling
//
// It does NOT modify Aether itself.

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Nibir1/Aether/aether"
)

// BatchOptions controls the concurrency of our local batch implementation.
type BatchOptions struct {
	Concurrency int
}

// BatchResult is a single fetch result.
type BatchResult struct {
	URL        string
	StatusCode int
	Body       []byte
	Err        error
}

// BatchResponse aggregates all results.
type BatchResponse struct {
	Results []BatchResult
}

// runBatchFetch executes N concurrent Fetch operations using the Aether client.
func runBatchFetch(ctx context.Context, cli *aether.Client, urls []string, opts BatchOptions) (*BatchResponse, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 4
	}

	sem := make(chan struct{}, opts.Concurrency)
	results := make([]BatchResult, len(urls))

	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)

		i := i
		u := u

		go func() {
			defer wg.Done()

			sem <- struct{}{}        // acquire slot
			defer func() { <-sem }() // release

			resp, err := cli.Fetch(ctx, u)
			if err != nil {
				results[i] = BatchResult{
					URL: u,
					Err: err,
				}
				return
			}

			results[i] = BatchResult{
				URL:        resp.URL,
				StatusCode: resp.StatusCode,
				Body:       resp.Body,
				Err:        nil,
			}
		}()
	}

	wg.Wait()
	return &BatchResponse{Results: results}, nil
}

func main() {
	cli, err := aether.NewClient(
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println("=== Test : Batch Fetch (Standalone Wrapper) ===")

	urls := []string{
		"https://example.com/",
		"https://www.wikipedia.org/",
		"https://news.ycombinator.com/",
	}

	batch, err := runBatchFetch(ctx, cli, urls, BatchOptions{
		Concurrency: 2,
	})
	if err != nil {
		log.Fatalf("Batch error: %v", err)
	}

	for i, r := range batch.Results {
		fmt.Printf("Result %d:\n", i+1)
		fmt.Printf("  URL: %s\n", r.URL)

		if r.Err != nil {
			fmt.Printf("  ERROR: %v\n\n", r.Err)
			continue
		}

		fmt.Printf("  StatusCode: %d\n", r.StatusCode)
		fmt.Printf("  Body len:   %d\n\n", len(r.Body))
	}
}
