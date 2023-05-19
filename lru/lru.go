package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	cap  int64
	size int64
	ll   *list.List
	eles map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type entry struct {
	key string
	val Value
}

// Value use Len to count how many size it takes.
type Value interface {
	Len() int
}

// New is the Constructor of Cache. cap = 0 means no limit.
func New(cap int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		cap:       cap,
		ll:        list.New(),
		eles:      make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look ups a key's val.
func (c *Cache) Get(key string) (val Value, ok bool) {
	if ele, ok := c.eles[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		return kv.val, true
	}
	return
}

// Eviction removes the oldest item.
func (c *Cache) Eviction() {
	ele := c.ll.Front()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.eles, kv.key)
		c.size -= int64(len(kv.key)) + int64(kv.val.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.val)
		}
	}
}

// Add adds a val to the eles.
func (c *Cache) Add(key string, val Value) {
	if e, ok := c.eles[key]; ok {
		c.ll.MoveToBack(e)
		kv := e.Value.(*entry)
		c.size += int64(val.Len()) - int64(kv.val.Len())
		kv.val = val
	} else {
		ele := c.ll.PushBack(&entry{key, val})
		c.eles[key] = ele
		c.size += int64(len(key)) + int64(val.Len())
	}
	for c.cap != 0 && c.cap < c.size {
		c.Eviction()
	}
}

// Len the number of eles entries.
func (c *Cache) Len() int {
	return c.ll.Len()
}
