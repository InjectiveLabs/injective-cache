package cache

import "errors"

var (
	ErrCacheMiss            = errors.New("not found in cache")
	ErrInvalidKey           = errors.New("key is invalid")
	ErrInvalidValue         = errors.New("value is invalid")
	ErrMissingFetchFunction = errors.New("missing fetch function")
)
