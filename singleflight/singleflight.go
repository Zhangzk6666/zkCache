package singleflight

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val []byte
	err error
}

// 当缓存中不存在某一key时，此时对该key的访问都将打到数据库中，且这些访问并发执行 => 可以将这些访问视为【一组】访问。
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// 并发获取同一个key，除了第一个其他阻塞（当第一个ing时）
func (g *Group) Do(key string, code int64, fn func() ([]byte, error)) ([]byte, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}
