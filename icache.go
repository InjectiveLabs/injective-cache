package cache

import (
	"errors"
)

type Cache interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error

	BatchGet(keys ...string) ([][]byte, error)
	// key1, value1, key2, value2 ...
	// string, bytes, string,bytes
	BatchSet(keyvalues ...interface{}) error
	Close() error
}

var ErrCacheMiss = errors.New("not found in cache")
