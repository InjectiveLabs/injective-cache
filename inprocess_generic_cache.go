package cache

import (
	"context"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

type MemLibCache struct {
	// caveat:
	// libcache won't evict upon timeout, eviction only happens when it's expired + the cache key is read
	cache libcache.Cache
}

func NewMemLibCache(cap int, ttl time.Duration) *MemLibCache {
	c := libcache.LRU.New(cap) // new thread-safe cache
	c.SetTTL(ttl)
	return &MemLibCache{
		cache: c,
	}
}

func (l *MemLibCache) SetWithTTL(_ context.Context, key interface{}, value interface{}, ttl time.Duration) (err error) {
	l.cache.StoreWithTTL(key, value, ttl)
	return nil
}

func (l *MemLibCache) Set(_ context.Context, key interface{}, value interface{}) (err error) {
	// with default ttl
	l.cache.Store(key, value)
	return nil
}

func (l *MemLibCache) Get(_ context.Context, key interface{}) (interface{}, error) {
	value, exist := l.cache.Load(key)
	if !exist {
		return nil, ErrCacheMiss
	}
	return value, nil
}

func (l *MemLibCache) Del(_ context.Context, key interface{}) (err error) {
	l.cache.Delete(key)
	return nil
}
