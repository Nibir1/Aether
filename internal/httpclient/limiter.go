// internal/httpclient/limiter.go
//
// This file implements simple concurrency limiting for outbound HTTP
// requests, both globally and per-host. It helps ensure that Aether
// behaves politely when accessing remote servers.
package httpclient

import (
	"context"
	"sync"
)

// hostLimiter controls concurrent access to remote hosts.
type hostLimiter struct {
	globalCh chan struct{}
	maxPer   int

	mu      sync.Mutex
	perHost map[string]chan struct{}
}

// newHostLimiter constructs a limiter with the given global and
// per-host concurrency limits.
func newHostLimiter(maxHosts, maxPerHost int) *hostLimiter {
	if maxHosts <= 0 {
		maxHosts = 4
	}
	if maxPerHost <= 0 {
		maxPerHost = 4
	}
	return &hostLimiter{
		globalCh: make(chan struct{}, maxHosts),
		maxPer:   maxPerHost,
		perHost:  make(map[string]chan struct{}),
	}
}

// Acquire reserves a slot for the given host. It respects context
// cancellation.
func (l *hostLimiter) Acquire(ctx context.Context, host string) error {
	select {
	case l.globalCh <- struct{}{}:
		// acquired global slot
	case <-ctx.Done():
		return ctx.Err()
	}

	ch := l.getHostChan(host)

	select {
	case ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		<-l.globalCh // release global
		return ctx.Err()
	}
}

// Release frees the slot for the given host.
func (l *hostLimiter) Release(host string) {
	l.mu.Lock()
	ch, ok := l.perHost[host]
	l.mu.Unlock()

	if ok {
		select {
		case <-ch:
		default:
			// should not happen, but avoid panic
		}
	}

	select {
	case <-l.globalCh:
	default:
	}
}

func (l *hostLimiter) getHostChan(host string) chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch, ok := l.perHost[host]
	if !ok {
		ch = make(chan struct{}, l.maxPer)
		l.perHost[host] = ch
	}
	return ch
}
