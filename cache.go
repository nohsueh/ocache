package ocache

import (
	"github.com/nohsueh/ocache/lru"
	"sync"
)

type Cache struct {
	mu    sync.Mutex
	cache *lru.Cache
	bytes int64
}

func (c *Cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = lru.New(c.bytes, nil)
	}
	c.cache.Add(key, value)
}

func (c *Cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return
	}

	if v, ok := c.cache.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
