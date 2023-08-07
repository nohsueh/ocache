package ocache

import (
	"github.com/nohsueh/ocache/lru"
	"sync"
)

type Cache struct {
	mu    sync.Mutex
	cache *lru.Cache
	cap   int64
}

func (c *Cache) add(key string, view ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = lru.New(c.cap, nil)
	}
	c.cache.Add(key, view)
}

func (c *Cache) get(key string) (view ByteView, ok bool) {
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
