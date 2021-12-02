package httpcache

import (
	"time"

	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"github.com/hashicorp/golang-lru"
)

// MemoryCache is an implemtation of Cache that stores responses in an in-memory map.
type MemoryCache struct {
	cache  *lru.Cache // key:string  value:Item
	expire time.Duration
}

type Item struct {
	resp []byte
	t    time.Time
}

// Get returns the []byte representation of the response and true if present, false if not
func (c *MemoryCache) Get(key string) (resp []byte, ok bool) {
	if value, ok := c.cache.Get(key); ok {
		if item, ok := value.(Item); ok {
			if c.expire > 0 && time.Since(item.t) < c.expire {
				return item.resp, true
			}
		}
	}
	return
}

// Set saves response resp to the cache with key
func (c *MemoryCache) Set(key string, resp []byte) {
	c.cache.Add(key, Item{resp: resp, t: time.Now()})
}

// Delete removes key from the cache
func (c *MemoryCache) Delete(key string) {
	c.cache.Remove(key)
}

// NewMemoryCache returns a new Cache that will store items in an in-memory map
func NewMemoryCache(size int, expire time.Duration) (*MemoryCache, error) {
	if size <= 0 {
		size = DefaultCacheSize
	}
	cache, err := lru.New(size)
	if err != nil { // 只有size <= 0才会报错
		return nil, errors.WithStack(err)
	}

	return &MemoryCache{
		cache:  cache,
		expire: expire,
	}, nil
}
