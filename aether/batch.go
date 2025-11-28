// aether/batch.go
//
// Batch subsystem (Stage 16).
//
// Provides a concurrency-friendly wrapper that fetches multiple URLs using
// Aether’s robots.txt-compliant internal fetch pipeline.
//
// Guarantees:
//   • robots.txt compliance (via internal httpclient)
//   • per-host fairness + rate limiting
//   • automatic gzip/deflate decoding
//   • stable per-URL errors rather than global failures
//   • ordering preserved exactly as input.
//
// Future extensions may incorporate plugins, TOON encoding, or transform passes.

package aether

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// BatchOptions configures batch fetch behavior.
type BatchOptions struct {
	// Maximum concurrent workers. If <= 0, defaults to 4.
	Concurrency int

	// Optional extra headers applied to every fetch.
	Headers http.Header
}

// BatchItemResult is the result of one fetch operation.
type BatchItemResult struct {
	URL        string
	StatusCode int
	Body       []byte
	Header     http.Header
	Err        error
}

// BatchResult contains all results in the same order as input.
type BatchResult struct {
	Results []BatchItemResult
}

// ErrNilClient indicates methods were called on a nil *Client.
var ErrNilClient = errors.New("aether: nil client")

// Batch performs robots-compliant GET operations on multiple URLs.
//
// It always returns a *BatchResult, even if some items fail.
// The error return only signals fatal client misconfiguration (nil client/globals).
//
// Per-item errors are stored inside Results[i].Err.
func (c *Client) Batch(ctx context.Context, urls []string, opts BatchOptions) (*BatchResult, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	if c.fetcher == nil {
		return nil, fmt.Errorf("aether: client fetcher is not initialized")
	}

	n := len(urls)
	if n == 0 {
		return &BatchResult{Results: nil}, nil
	}

	// Worker pool size
	workers := opts.Concurrency
	if workers <= 0 {
		workers = 4
	}

	// Preallocate results in correct index order
	results := make([]BatchItemResult, n)

	// Immutable header clone (avoid shared map mutation)
	var hdr http.Header
	if opts.Headers != nil {
		hdr = opts.Headers.Clone()
	}

	// Job descriptor
	type job struct {
		idx int
		url string
	}

	jobs := make(chan job)
	wg := sync.WaitGroup{}
	wg.Add(workers)

	// Worker function
	worker := func() {
		defer wg.Done()

		for j := range jobs {
			url := strings.TrimSpace(j.url)
			res := BatchItemResult{URL: url}

			if url == "" {
				res.Err = fmt.Errorf("aether: empty URL")
				results[j.idx] = res
				continue
			}

			// Honor context cancellation before fetch
			select {
			case <-ctx.Done():
				res.Err = ctx.Err()
				results[j.idx] = res
				continue
			default:
			}

			resp, err := c.fetcher.Fetch(ctx, url, hdr)
			if err != nil {
				res.Err = err
				results[j.idx] = res
				continue
			}

			res.StatusCode = resp.StatusCode
			res.Body = resp.Body

			if resp.Header != nil {
				res.Header = resp.Header.Clone()
			} else {
				res.Header = http.Header{}
			}

			results[j.idx] = res
		}
	}

	// Start workers
	for w := 0; w < workers; w++ {
		go worker()
	}

	// Feed jobs
	go func() {
		defer close(jobs)
		for i, u := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- job{idx: i, url: u}:
			}
		}
	}()

	// Wait for workers
	wg.Wait()

	return &BatchResult{Results: results}, nil
}
