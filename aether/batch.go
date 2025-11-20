// aether/batch.go
//
// Batch subsystem (Stage 16).
//
// This provides a high-level, concurrency-friendly wrapper that fetches
// multiple URLs at once using Aether’s internal fetch pipeline.
//
// Features:
//   • robots.txt-compliant (via underlying http fetcher)
//   • per-host fairness & rate-limiting automatically preserved
//   • configurable concurrency
//   • returns structured BatchResult with per-URL errors
//
// Future expansions may add plugin hooks, transform passes, or TOON/JSON
// serialization helpers.

package aether

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

// BatchOptions configures the behavior of batch fetch operations.
type BatchOptions struct {
	// Maximum concurrent workers. If <= 0, defaults to 4.
	Concurrency int

	// Optional additional headers applied to each fetch request.
	Headers http.Header
}

// BatchItemResult represents the outcome of a single fetch in the batch.
type BatchItemResult struct {
	URL        string
	StatusCode int
	Body       []byte
	Header     http.Header
	Err        error
}

// BatchResult contains results for all fetched URLs in the same order
// as the input slice.
type BatchResult struct {
	Results []BatchItemResult
}

// ErrNilClient is returned when methods are called on a nil *Client receiver.
// This prevents panics and provides a predictable error signal.
var ErrNilClient = errors.New("aether: nil client")

// Batch fetches a list of URLs using Aether’s internal robots.txt-compliant
// fetcher and returns structured results. The function preserves input
// ordering in the result set.
func (c *Client) Batch(ctx context.Context, urls []string, opts BatchOptions) (*BatchResult, error) {
	if c == nil {
		return nil, ErrNilClient
	}

	n := len(urls)
	if n == 0 {
		return &BatchResult{Results: nil}, nil
	}

	// Default concurrency
	workers := opts.Concurrency
	if workers <= 0 {
		workers = 4
	}

	// The output slice is pre-allocated in correct order.
	results := make([]BatchItemResult, n)

	// A channel of jobs: each job is (index, url)
	type job struct {
		idx int
		url string
	}

	jobs := make(chan job)
	wg := sync.WaitGroup{}

	// Worker function using Aether.fetcher
	worker := func() {
		defer wg.Done()

		for j := range jobs {
			res := BatchItemResult{URL: j.url}

			resp, err := c.fetcher.Fetch(ctx, j.url, opts.Headers)
			if err != nil {
				res.Err = err
				results[j.idx] = res
				continue
			}

			res.Body = resp.Body
			res.Header = resp.Header.Clone()
			res.StatusCode = resp.StatusCode

			results[j.idx] = res
		}
	}

	// Start workers
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go worker()
	}

	// Feed the jobs
	go func() {
		for i, u := range urls {
			select {
			case <-ctx.Done():
				// If context canceled, close and exit workers early.
				close(jobs)
				return
			case jobs <- job{idx: i, url: u}:
			}
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	wg.Wait()

	return &BatchResult{Results: results}, nil
}
