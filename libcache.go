package cache

import (
	"context"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

var _ SimpleCache = (*TypedLibCache[string])(nil)

type TypedLibCache[T any] struct {
	// caveat:
	// libcache won't evict upon timeout, eviction only happens when it's expired + the cache key is read
	cache libcache.Cache
}

func NewTypedLibCache[T any](cache libcache.Cache, ttl time.Duration) *TypedLibCache[T] {
	cache.SetTTL(ttl)
	return &TypedLibCache[T]{
		cache: cache,
	}
}

func (l *TypedLibCache[T]) SetWithTTL(_ context.Context, key interface{}, value interface{}, ttl time.Duration) (err error) {
	if _, ok := value.(T); !ok {
		return ErrInvalidValue
	}
	l.cache.StoreWithTTL(key, value, ttl)
	return nil
}

func (l *TypedLibCache[T]) Set(_ context.Context, key interface{}, value interface{}) (err error) {
	if _, ok := value.(T); !ok {
		return ErrInvalidValue
	}
	l.cache.Store(key, value)
	return nil
}

func (l *TypedLibCache[T]) Get(_ context.Context, key interface{}, value interface{}) (err error) {
	v, exist := l.cache.Load(key)
	if !exist {
		return ErrCacheMiss
	}
	if _, ok := value.(*T); !ok {
		return ErrInvalidValue
	}
	*value.(*T) = v.(T)
	return nil
}

func (l *TypedLibCache[T]) Del(_ context.Context, keys ...interface{}) (err error) {
	for _, key := range keys {
		l.cache.Delete(key)
	}
	return nil
}

func (l *TypedLibCache[T]) Clear(_ context.Context) (err error) {
	l.cache.Purge()
	return nil
}
