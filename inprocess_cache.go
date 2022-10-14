package cache

import (
	"errors"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

type memLibCache struct {
	// caveat:
	// libcache won't evict upon timeout, eviction only happens when it's expired + the cache key is read
	cache libcache.Cache
	Cache
}

func NewLibcache(cap int, ttl time.Duration) (Cache, error) {
	c := libcache.LRU.New(cap) // new thread-safe cache
	c.SetTTL(ttl)
	return &memLibCache{
		cache: c,
	}, nil
}

func (l *memLibCache) Set(key string, value []byte) (err error) {
	// with default ttl
	l.cache.Store(key, value)
	return nil
}

func (l *memLibCache) Get(key string) ([]byte, error) {
	value, exist := l.cache.Load(key)
	if !exist {
		return nil, ErrCacheMiss
	}
	return value.([]byte), nil
}

func (l *memLibCache) Del(key string) (err error) {
	l.cache.Delete(key)
	return nil
}

func (l *memLibCache) BatchGet(keys ...string) (result [][]byte, err error) {
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

func (l *memLibCache) BatchSet(keyvalues ...interface{}) error {
	if len(keyvalues)%2 != 0 {
		return errors.New("keyvalues len must be even")
	}

	for i := 0; i < len(keyvalues); i += 2 {
		key := keyvalues[i].(string)
		value := keyvalues[i+1].([]byte)
		l.cache.Store(key, value)
	}
	return nil
}

func (l *memLibCache) Close() error {
	l.cache.Purge()
	return nil
}
