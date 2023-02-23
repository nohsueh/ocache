package ocache

import (
	"fmt"
	pb "github.com/nohsueh/ocache/ocachepb"
	"github.com/nohsueh/ocache/singleflight"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function.
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Class is a Cache namespace and associated data loaded spread over.
type Class struct {
	name      string
	getter    Getter
	mainCache Cache
	peers     PeerPicker
	// use singleflight.Class to make sure that
	// each key is only fetched once
	loader *singleflight.Class
}

var (
	mu      sync.RWMutex
	classes = make(map[string]*Class)
)

// NewClass create a new instance of Class.
func NewClass(name string, cacheBytes int64, getter Getter) *Class {
	mu.Lock()
	defer mu.Unlock()
	if getter == nil {
		panic("nil Getter")
	}
	c := &Class{
		name:      name,
		getter:    getter,
		mainCache: Cache{bytes: cacheBytes},
		loader:    &singleflight.Class{},
	}
	classes[name] = c
	return c
}

// GetClass returns the named class previously created with NewClass, or nil if
// there's no such class.
func GetClass(name string) *Class {
	mu.RLock()
	defer mu.RUnlock()
	c := classes[name]
	return c
}

// Get b for a key from Cache.
func (c *Class) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if value, ok := c.mainCache.get(key); ok {
		log.Println("[Cache] hit")
		return value, nil
	}

	return c.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (c *Class) RegisterPeers(peers PeerPicker) {
	if c.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	c.peers = peers
}

func (c *Class) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	v, err := c.loader.Do(key,
		func() (interface{}, error) {
			if c.peers != nil {
				if peer, ok := c.peers.PickPeer(key); ok {
					if value, err = c.getFromPeer(peer, key); err == nil {
						return value, nil
					}
					log.Println("[GeeCache] Failed to get from peer", err)
				}
			}

			return c.getLocally(key)
		},
	)

	if err == nil {
		return v.(ByteView), nil
	}
	return
}

func (c *Class) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Class: c.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (c *Class) getLocally(key string) (ByteView, error) {
	bytes, err := c.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	c.populateCache(key, value)
	return value, nil
}

func (c *Class) populateCache(key string, value ByteView) {
	c.mainCache.add(key, value)
}
