package cache

import (
	"context"
	"errors"
)

type Cache interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error

	BatchGet(keys ...string) ([][]byte, error)
	// BatchSet
	// key1, value1, key2, value2 ...
	// string, bytes, string,bytes
	BatchSet(keyvalues ...interface{}) error
	Close() error
}

type CacheCtx interface {
	SetCtx(ctx context.Context, key string, value []byte) error
	GetCtx(ctx context.Context, key string) ([]byte, error)
	DelCtx(ctx context.Context, key string) error

	BatchGetCtx(ctx context.Context, keys ...string) ([][]byte, error)
	// BatchSetCtx
	// key1, value1, key2, value2 ...
	// string, bytes, string,bytes
	BatchSetCtx(ctx context.Context, keyvalues ...interface{}) error
	CloseCtx(ctx context.Context) error
}

var (
	ErrCacheMiss    = errors.New("not found in cache")
	ErrInvalidKey   = errors.New("key is invalid")
	ErrInvalidValue = errors.New("value is invalid")
)
