// internal/cache/cache.go
//
// This file defines Aether's internal caching abstraction. It supports
// multiple backends (memory, file, Redis) that can be composed together.
// High-level subsystems (Fetch, RSS, Search, OpenAPIs) interact only
// with this interface.

package cache

import (
	"time"

	"github.com/Nibir1/Aether/internal/log"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
}

type Config struct {
	MemoryEnabled bool
	MemoryTTL     time.Duration
	MemoryMax     int

	FileEnabled   bool
	FileTTL       time.Duration
	FileDirectory string

	RedisEnabled bool
	RedisTTL     time.Duration
	RedisAddress string

	Logger log.Logger
}

// compositeCache checks multiple caches in priority order:
// Memory → File → Redis → Miss.
type compositeCache struct {
	memory Cache
	file   Cache
	redis  Cache
	log    log.Logger
}

func NewComposite(cfg Config) Cache {
	var mem Cache
	if cfg.MemoryEnabled {
		mem = NewMemory(cfg.MemoryMax, cfg.MemoryTTL)
	}
	var file Cache
	if cfg.FileEnabled {
		file = NewFile(cfg.FileDirectory, cfg.FileTTL)
	}
	var redis Cache
	if cfg.RedisEnabled {
		redis = NewRedis(cfg.RedisAddress, cfg.RedisTTL)
	}

	return &compositeCache{
		memory: mem,
		file:   file,
		redis:  redis,
		log:    cfg.Logger,
	}
}

func (c *compositeCache) Get(key string) ([]byte, bool) {
	// 1) Memory
	if c.memory != nil {
		if v, ok := c.memory.Get(key); ok {
			c.log.Debugf("cache: memory hit %s", key)
			return v, true
		}
	}
	// 2) File
	if c.file != nil {
		if v, ok := c.file.Get(key); ok {
			c.log.Debugf("cache: file hit %s", key)
			if c.memory != nil {
				c.memory.Set(key, v, time.Hour) // promote
			}
			return v, true
		}
	}
	// 3) Redis
	if c.redis != nil {
		if v, ok := c.redis.Get(key); ok {
			c.log.Debugf("cache: redis hit %s", key)
			if c.memory != nil {
				c.memory.Set(key, v, time.Hour)
			}
			return v, true
		}
	}
	return nil, false
}

func (c *compositeCache) Set(key string, value []byte, ttl time.Duration) {
	if c.memory != nil {
		c.memory.Set(key, value, ttl)
	}
	if c.file != nil {
		c.file.Set(key, value, ttl)
	}
	if c.redis != nil {
		c.redis.Set(key, value, ttl)
	}
}
