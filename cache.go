package zkcache

import (
	"fmt"
	"sync"

	"github.com/Zhangzk6666/zkCache/lru"
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
func (c *synCache) getAll() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.GetAll()
}

func (c *synCache) set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Set(key, value)
}

func (c *synCache) remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Println("@@@@@@@@@!!!122211")

	c.lru.Remove(key)
}
