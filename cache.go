package zkcache

import (
	"sync"
	"zkCache/lru"
)

type synCache struct {
	mu  sync.Mutex
	lru *lru.Cache
}

func NewCache(maxSize int, onEvicted lru.OnEvictedFunc) *synCache {
	return &synCache{
		lru: lru.New(maxSize, onEvicted),
	}
}

func (c *synCache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Get(key)
}

func (c *synCache) set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Set(key, value)
}
