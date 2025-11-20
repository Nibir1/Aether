// internal/cache/memory.go
//
// Implements a simple thread-safe LRU memory cache with TTL.

package cache

import (
	"container/list"
	"sync"
	"time"
)

type memoryCache struct {
	max     int
	ttl     time.Duration
	mu      sync.Mutex
	ll      *list.List
	entries map[string]*list.Element
}

type entry struct {
	key     string
	value   []byte
	expires time.Time
}

func NewMemory(max int, ttl time.Duration) Cache {
	if max <= 0 {
		max = 128
	}
	if ttl <= 0 {
		ttl = 30 * time.Second
	}

	return &memoryCache{
		max:     max,
		ttl:     ttl,
		ll:      list.New(),
		entries: make(map[string]*list.Element),
	}
}

func (m *memoryCache) Get(key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ele, ok := m.entries[key]
	if !ok {
		return nil, false
	}

	ent := ele.Value.(*entry)
	if time.Now().After(ent.expires) {
		m.removeElement(ele)
		return nil, false
	}

	// LRU promotion
	m.ll.MoveToFront(ele)

	// Return a copy
	cpy := make([]byte, len(ent.value))
	copy(cpy, ent.value)
	return cpy, true
}

func (m *memoryCache) Set(key string, value []byte, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ttl <= 0 {
		ttl = m.ttl
	}

	if ele, ok := m.entries[key]; ok {
		m.ll.MoveToFront(ele)
		ent := ele.Value.(*entry)
		ent.value = cloneBytes(value)
		ent.expires = time.Now().Add(ttl)
		return
	}

	ent := &entry{
		key:     key,
		value:   cloneBytes(value),
		expires: time.Now().Add(ttl),
	}

	ele := m.ll.PushFront(ent)
	m.entries[key] = ele

	if m.ll.Len() > m.max {
		m.removeOldest()
	}
}

func (m *memoryCache) removeOldest() {
	ele := m.ll.Back()
	if ele != nil {
		m.removeElement(ele)
	}
}

func (m *memoryCache) removeElement(ele *list.Element) {
	m.ll.Remove(ele)
	ent := ele.Value.(*entry)
	delete(m.entries, ent.key)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
