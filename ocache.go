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

// A Relation is a Cache namespace and associated data loaded spread over.
type Relation struct {
	name   string
	getter Getter
	cache  Cache
	peers  PeerPicker
	// use singleflight.Relation to make sure that
	// each key is only fetched once
	loader *singleflight.Relation
}

var (
	mu        sync.RWMutex
	relations = make(map[string]*Relation)
)

// NewRelation create a new instance of Relation.
func NewRelation(name string, cacheBytes int64, getter Getter) *Relation {
	mu.Lock()
	defer mu.Unlock()
	if getter == nil {
		panic("nil Getter")
	}
	r := &Relation{
		name:   name,
		getter: getter,
		cache:  Cache{cap: cacheBytes},
		loader: &singleflight.Relation{},
	}
	relations[name] = r
	return r
}

// GetRelation returns the named class previously created with NewRelation, or nil if
// there's no such class.
func GetRelation(name string) *Relation {
	mu.RLock()
	defer mu.RUnlock()
	r := relations[name]
	return r
}

// Get bytes for a key from Cache.
func (r *Relation) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if view, ok := r.cache.get(key); ok {
		log.Println("[Cache] hit")
		return view, nil
	}

	return r.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (r *Relation) RegisterPeers(peers PeerPicker) {
	if r.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	r.peers = peers
}

func (r *Relation) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	v, err := r.loader.Do(key,
		func() (interface{}, error) {
			if r.peers != nil {
				if peer, ok := r.peers.PickPeer(key); ok {
					if value, err = r.getFromPeer(peer, key); err == nil {
						return value, nil
					}
					log.Println("[GeeCache] Failed to get from peer", err)
				}
			}

			return r.getLocally(key)
		},
	)

	if err == nil {
		return v.(ByteView), nil
	}
	return
}

func (r *Relation) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Relation: r.name,
		Key:      key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes: res.Value}, nil
}

func (r *Relation) getLocally(key string) (ByteView, error) {
	bytes, err := r.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{bytes: cloneBytes(bytes)}
	r.populateCache(key, value)
	return value, nil
}

func (r *Relation) populateCache(key string, value ByteView) {
	r.cache.add(key, value)
}
