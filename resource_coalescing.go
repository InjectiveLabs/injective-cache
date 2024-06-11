package cache

import (
	"context"
	"sync"
	"sync/atomic"
)

type CachedResourceCoalescing[K comparable, T any] struct {
	OnErr    func(error)
	cache    TTLCache
	cacheMX  sync.RWMutex
	inFlight map[K]*resource
	once     sync.Once
}

func NewCachedResourceCoalescing[K comparable, T any](cache TTLCache) *CachedResourceCoalescing[K, T] {
	return &CachedResourceCoalescing[K, T]{
		cache:    cache,
		inFlight: make(map[K]*resource),
	}
}

type resource struct {
	value   interface{}
	err     error
	waiting atomic.Int64
	done    chan struct{}
}

// GetOnceWithStore is a helper function that fetches a value from the cache or executes the provided function to get the value.
func GetOnceWithStore[K comparable, T any](
	ctx context.Context,
	crc *CachedResourceCoalescing[K, T],
	key K,
	get func() (result T, storeInCache bool, err error),
) (result T, err error) {
	var cachedRes T
	cacheErr := crc.cache.Get(ctx, key, &cachedRes)
	if cacheErr == nil {
		return cachedRes, nil
	}

	crc.cacheMX.Lock()
	res, found := crc.inFlight[key]
	if found {
		crc.inFlight[key].waiting.Add(1)
		defer crc.inFlight[key].waiting.Add(-1)
		crc.cacheMX.Unlock()
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-res.done:
			return res.value.(T), res.err
		}
	}

	// not found, create a new promise and query the result
	res = &resource{
		done: make(chan struct{}),
	}
	crc.inFlight[key] = res
	crc.cacheMX.Unlock()

	// execute the function
	var store bool
	result, store, err = get()
	res.value = result
	res.err = err
	close(res.done)

	// store the result in the cache if needed
	if store {
		if setErr := crc.cache.Set(ctx, key, result); setErr != nil && crc.OnErr != nil {
			crc.OnErr(setErr)
		}
	}

	// remove the promise from the in-flight map
	crc.cacheMX.Lock()
	delete(crc.inFlight, key)
	crc.cacheMX.Unlock()

	return result, err
}

// GetOnce is a helper function that fetches a value from the cache or executes the provided function to get the value.
//
//	It will store the value in the cache if the function returns no error
func GetOnce[K comparable, T any](ctx context.Context, crc *CachedResourceCoalescing[K, T], key K, execute func() (T, error)) (T, error) {
	return GetOnceWithStore(ctx, crc, key, func() (T, bool, error) {
		v, executeErr := execute()
		return v, executeErr == nil, executeErr
	})
}
