package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32.
type Hash func(data []byte) uint32

// Map contains all hashed keys.
type Map struct {
	hash     Hash
	replicas int
	keys     []int // Sorted
	hashes   map[int]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}

	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashes:   make(map[int]string),
	}

	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, k := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(k + strconv.Itoa(i))))
			m.keys = append(m.keys, hash)
			m.hashes[hash] = k
		}
	}

	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	h := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	i := sort.Search(len(m.keys),
		func(i int) bool {
			return m.keys[i] >= h
		},
	)

	return m.hashes[m.keys[i%len(m.keys)]]
}
