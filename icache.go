package cache

import (
	"context"
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
	Close() error
}
