package lru

import (
	"container/list"
)

type Cache struct {
	// 允许最大内存空间，0表示不限制  单位字节
	maxSize int
	// 当前占用内存空间 单位字节
	size      int
	list      *list.List
	cache     map[string]*list.Element
	OnEvicted OnEvictedFunc
}

type OnEvictedFunc func(key string, value string)

type entry struct {
	key   string
	value string
}

func New(maxSize int, onEvicted OnEvictedFunc) *Cache {
	return &Cache{
		maxSize:   maxSize,
		size:      0,
		list:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 缓存个数
func (c *Cache) Len() int {
	return c.list.Len()
}

func (c *Cache) Get(key string) (value string, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}
func (c *Cache) GetAll() map[string]string {
	copy := make(map[string]string)
	for k, v := range c.cache {
		copy[k] = v.Value.(*entry).value
	}
	return copy
}

func (c *Cache) Remove(key string) {
	if _, ok := c.Get(key); ok {
		delete(c.cache, key)
	}
}

func (c *Cache) Set(key string, value string) {
	if ele, ok := c.cache[key]; ok {
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.size += (len(value) - len(kv.value))
		kv.value = value
	}

	ele := c.list.PushFront(&entry{
		key:   key,
		value: value,
	})
	c.cache[key] = ele
	c.size += (len(key) + len(value))
	// 新添加 || 更新 都有可能触发
	for c.maxSize != 0 && c.maxSize < c.size {
		c.removeBack()
	}
}

func (c *Cache) removeBack() {
	if ele := c.list.Back(); ele != nil {
		c.list.Remove(ele)
		kv := ele.Value.(*entry)
		c.size -= (len(kv.key) + len(kv.value))
		delete(c.cache, kv.key)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}
