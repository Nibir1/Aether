// internal/crawl/queue.go
//
// Package crawl implements Aether's legal, robots.txt-compliant crawl engine.
// This file defines the frontier queue used by the crawler to manage URLs
// pending visitation.
//
// The queue is a simple FIFO structure storing (URL, depth) pairs. It is
// safe for concurrent use by the crawler workers via a mutex, but it is
// deliberately minimal: there is no blocking or condition variable logic
// here. The crawler orchestrator controls when to poll or stop.

package crawl

import "sync"

// FrontierItem represents a single entry in the crawl frontier.
// Depth is measured from the starting URL (depth 0).
type FrontierItem struct {
	URL   string
	Depth int
}

// FrontierQueue is a thread-safe FIFO queue of FrontierItem values.
//
// The queue is used by the crawler to schedule which URLs to visit next.
// It does not perform any URL normalization or filtering; those concerns
// are handled by higher-level components (rules, visit map, etc.).
type FrontierQueue struct {
	mu    sync.Mutex
	items []FrontierItem
}

// NewFrontierQueue constructs an empty frontier queue.
func NewFrontierQueue() *FrontierQueue {
	return &FrontierQueue{
		items: make([]FrontierItem, 0),
	}
}

// Enqueue adds a new item to the end of the queue.
// It is safe to call from multiple goroutines.
func (q *FrontierQueue) Enqueue(item FrontierItem) {
	if item.URL == "" {
		// silently ignore empty URLs; the crawler performs more
		// thorough validation before enqueuing in normal operation.
		return
	}

	q.mu.Lock()
	q.items = append(q.items, item)
	q.mu.Unlock()
}

// Dequeue removes and returns the oldest item from the queue.
// The boolean return value is false if the queue is empty.
func (q *FrontierQueue) Dequeue() (FrontierItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return FrontierItem{}, false
	}

	it := q.items[0]
	// shift slice
	copy(q.items[0:], q.items[1:])
	q.items = q.items[:len(q.items)-1]

	return it, true
}

// Len returns the current number of items in the queue.
//
// This is primarily useful for monitoring and unit tests; the crawler
// uses Dequeue's boolean return to detect emptiness.
func (q *FrontierQueue) Len() int {
	q.mu.Lock()
	n := len(q.items)
	q.mu.Unlock()
	return n
}

// Empty reports whether the queue is currently empty.
func (q *FrontierQueue) Empty() bool {
	return q.Len() == 0
}
