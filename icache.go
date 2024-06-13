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

// TTLCache is a cache interface that allows any type of key and value
// as long as the key is hashable and the value is serializable.
//
//go:generate mockgen -destination=icache_mock.go -package=cache github.com/InjectiveLabs/injective-cache TTLCache
type TTLCache interface {
	Set(ctx context.Context, key any, value any) (err error)
	SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) (err error)
	Get(ctx context.Context, key any, value any) (err error)
	Del(ctx context.Context, keys ...any) (err error)
	Clear(ctx context.Context) (err error)
}

// Get is a generic helper to retrieve any type of value from a TTLCache.
func Get[T any](ctx context.Context, c TTLCache, key any) (value T, err error) {
	return value, c.Get(ctx, key, &value)
}
