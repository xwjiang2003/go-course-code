package cache

import "sync"

// Cache 是并发安全的内存缓存（演示简化版）。
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *Cache) Set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
}
