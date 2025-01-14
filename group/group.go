package group

import (
	"GeeCache/byteview"
	"GeeCache/cache"
	"fmt"
	"sync"
)

// Getter loads data for a key
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc implements Getter with a function
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group is a cache namespace
type Group struct {
	name      string
	getter    Getter
	mainCache cache.Cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) RegisterPeers(peers PeerPicker) {
	g.peers = peers
}

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache.Cache{CacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (cache.ByteView, error) {
	if key == "" {
		return byteview.ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.Get(key); ok {
		return v, nil
	}

	return g.load(key)
}

// load loads key by calling getLocally or getFromPeer
func (g *Group) load(key string) (value byteview.ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (byteview.ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return byteview.ByteView{}, err
	}
	value := byteview.ByteView{B: bytes}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value byteview.ByteView) {
	g.mainCache.Add(key, value)
}

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
