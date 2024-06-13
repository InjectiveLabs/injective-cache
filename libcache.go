package cache

import (
	"context"
	"time"

	"github.com/shaj13/libcache"
)

var _ TTLCache = (*TypedLibCache[string, string])(nil)

type TypedLibCache[K comparable, T any] struct {
	// caveat:
	// libcache won't evict upon timeout, eviction only happens when it's expired + the cache key is read
	cache libcache.Cache
}

func NewTypedLibCache[K comparable, T any](cache libcache.Cache, ttl time.Duration) *TypedLibCache[K, T] {
	cache.SetTTL(ttl)
	return &TypedLibCache[K, T]{
		cache: cache,
	}
}

func (l *TypedLibCache[K, T]) SetWithTTL(_ context.Context, key any, value any, ttl time.Duration) (err error) {
	if _, ok := value.(T); !ok {
		return ErrInvalidValue
	}
	l.cache.StoreWithTTL(key, value, ttl)
	return nil
}

func (l *TypedLibCache[K, T]) Set(_ context.Context, key any, value any) (err error) {
	if _, ok := key.(K); !ok {
		return ErrInvalidKey
	}
	if _, ok := value.(T); !ok {
		return ErrInvalidValue
	}
	l.cache.Store(key, value)
	return nil
}

func (l *TypedLibCache[K, T]) Get(_ context.Context, key any, value any) (err error) {
	if _, ok := key.(K); !ok {
		return ErrInvalidKey
	}
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

func (l *TypedLibCache[K, T]) Del(_ context.Context, keys ...any) (err error) {
	for _, key := range keys {
		if _, ok := key.(K); !ok {
			return ErrInvalidKey
		}
		l.cache.Delete(key)
	}
	return nil
}

func (l *TypedLibCache[K, T]) Clear(_ context.Context) (err error) {
	l.cache.Purge()
	return nil
}
