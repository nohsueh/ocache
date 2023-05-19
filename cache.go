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

func (f *Cache) add(key string, view ByteView) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.cache == nil {
		f.cache = lru.New(f.cap, nil)
	}
	f.cache.Add(key, view)
}

func (f *Cache) get(key string) (view ByteView, ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.cache == nil {
		return
	}

	if v, ok := f.cache.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
