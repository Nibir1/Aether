// internal/crawl/depth.go
//
// This file implements depth tracking utilities for Aether's crawl subsystem.
// The crawler assigns a "depth" value to each URL, beginning with the root
// URL at depth 0. Each discovered link increments depth by +1.
//
// Depth rules serve two purposes:
//   1. To prevent infinite crawling of large sites.
//   2. To allow callers to specify how many link-levels to traverse.
//
// Depth logic is intentionally isolated in this small file so it can be
// easily unit-tested and reused by the crawler engine.
//
// Note:
// The crawler orchestrator is responsible for ensuring that only valid
// (URL, depth) pairs are enqueued in the frontier. This file only provides
// helper utilities.

package crawl

// DepthLimit is a simple struct used for validating depth transitions.
type DepthLimit struct {
	MaxDepth int // maximum allowed depth (0-based). If MaxDepth < 0, unlimited.
}

// NewDepthLimit constructs a new DepthLimit.
//
// If maxDepth < 0, the crawler treats depth as unlimited.
// Depth 0 = root URL.
// Depth 1 = root's outgoing links.
// Depth 2 = links from depth 1 pages, etc.
func NewDepthLimit(maxDepth int) DepthLimit {
	return DepthLimit{MaxDepth: maxDepth}
}

// Allowed reports whether a page at `depth` is allowed to be visited
// according to the configured max depth.
//
// If MaxDepth < 0, depth is unlimited.
func (d DepthLimit) Allowed(depth int) bool {
	if depth < 0 {
		return false
	}
	if d.MaxDepth < 0 {
		return true
	}
	return depth <= d.MaxDepth
}

// Next returns the next depth for child links.
//
// Parents at depth N produce children at depth N+1.
func (d DepthLimit) Next(parentDepth int) int {
	return parentDepth + 1
}

// Exceeded reports whether this (parentDepth + 1) exceeds MaxDepth.
//
// Useful when deciding whether to enqueue outgoing links.
func (d DepthLimit) Exceeded(parentDepth int) bool {
	next := parentDepth + 1
	return !d.Allowed(next)
}
