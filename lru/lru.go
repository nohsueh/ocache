package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	cap      int64
	size     int64
	list     *list.List
	cacheMap map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value use Len to count how many size it takes.
type Value interface {
	Len() int
}

// New is the Constructor of Cache. cap = 0 means no limit.
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		cap:       maxBytes,
		list:      list.New(),
		cacheMap:  make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look ups a key's value.
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cacheMap[key]; ok {
		c.list.MoveToBack(element)
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return
}

// Eviction removes the oldest item.
func (c *Cache) Eviction() {
	front := c.list.Front()
	if front != nil {
		c.list.Remove(front)
		kv := front.Value.(*entry)
		delete(c.cacheMap, kv.key)
		c.size -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cacheMap.
func (c *Cache) Add(key string, value Value) {
	if element, ok := c.cacheMap[key]; ok {
		c.list.MoveToBack(element)
		kv := element.Value.(*entry)
		c.size += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		front := c.list.PushBack(&entry{key, value})
		c.cacheMap[key] = front
		c.size += int64(len(key)) + int64(value.Len())
	}
	for c.cap != 0 && c.cap < c.size {
		c.Eviction()
	}
}

// Len the number of cacheMap entries.
func (c *Cache) Len() int {
	return c.list.Len()
}
