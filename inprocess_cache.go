package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

var _ Cache = (*memLibCache)(nil)

type memLibCache struct {
	// caveat:
	// libcache won't evict upon timeout, eviction only happens when it's expired + the cache key is read
	cache libcache.Cache
}

func NewLibcache(cap int, ttl time.Duration) (*memLibCache, error) {
	c := libcache.LRU.New(cap) // new thread-safe cache
	c.SetTTL(ttl)
	return &memLibCache{
		cache: c,
	}, nil
}

func (l *memLibCache) Set(ctx context.Context, key string, value []byte) (err error) {
	// with default ttl
	l.cache.Store(key, value)
	return nil
}

func (l *memLibCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, exist := l.cache.Load(key)
	if !exist {
		return nil, ErrCacheMiss
	}
	return value.([]byte), nil
}

func (l *memLibCache) Del(ctx context.Context, key string) (err error) {
	l.cache.Delete(key)
	return nil
}

func (l *memLibCache) BatchGet(ctx context.Context, keys ...string) (result [][]byte, err error) {
	for _, k := range keys {
		value, exist := l.cache.Load(k)
		if !exist {
			result = append(result, nil)
			continue
		}
		result = append(result, value.([]byte))
	}
	return result, nil
}

func (l *memLibCache) BatchSet(ctx context.Context, keyvalues ...interface{}) error {
	if len(keyvalues)%2 != 0 {
		return errors.New("keyvalues len must be even")
	}

	for i := 0; i < len(keyvalues); i += 2 {
		key, ok := keyvalues[i].(string)
		if !ok {
			return fmt.Errorf("%w at index %d: expected string, got %T", ErrInvalidKey, i, keyvalues[i])
		}

		value, ok := keyvalues[i+1].([]byte)
		if !ok {
			return fmt.Errorf("%w at index %d: expected []byte, got %T", ErrInvalidValue, i, keyvalues[i])
		}

		l.cache.Store(key, value)
	}
	return nil
}

func (l *memLibCache) Close() error {
	l.cache.Purge()
	return nil
}
