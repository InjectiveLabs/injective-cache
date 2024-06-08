package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error

	BatchGet(ctx context.Context, keys ...string) ([][]byte, error)
	// BatchSet
	// key1, value1, key2, value2 ...
	// string, bytes, string,bytes
	BatchSet(ctx context.Context, keyvalues ...interface{}) error

	IsRunning(ctx context.Context) bool

	Close() error
}

// SimpleCache is a cache interface that allows any type of key and value
// as long as the key is hashable and the value is serializable.
type SimpleCache interface {
	Set(ctx context.Context, key interface{}, value interface{}) (err error)
	SetWithTTL(ctx context.Context, key interface{}, value interface{}, ttl time.Duration) (err error)
	Get(ctx context.Context, key interface{}, value interface{}) (err error)
	Del(ctx context.Context, keys ...interface{}) (err error)
	Clear(ctx context.Context) (err error)
}

// Get is a generic helper to retrieve any type of value from a SimpleCache.
func Get[T any](ctx context.Context, c SimpleCache, key interface{}) (value T, err error) {
	return value, c.Get(ctx, key, &value)
}
