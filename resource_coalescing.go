package cache

import (
	"context"
	"sync"
	"sync/atomic"
)

// ResourceCoalescingCache is a cache that coalesces multiple requests for the same resource into a single request
// to prevent cache stampedes. It is useful when the resource is expensive to compute and can be shared among multiple
type ResourceCoalescingCache[K comparable, T any] struct {
	OnErr    func(error)
	cache    TTLCache
	cacheMX  sync.RWMutex
	inFlight map[K]*resource
	once     sync.Once
	fetch    func() (T, error)
}

// NewResourceCoalescingCache creates a new ResourceCoalescingCache
func NewResourceCoalescingCache[K comparable, T any](cache TTLCache, fetch func() (T, error)) *ResourceCoalescingCache[K, T] {
	if fetch == nil {
		fetch = func() (res T, err error) {
			return res, ErrMissingFetchFunction
		}
	}
	return &ResourceCoalescingCache[K, T]{
		cache:    cache,
		fetch:    fetch,
		inFlight: make(map[K]*resource),
	}
}

type resource struct {
	value   interface{}
	err     error
	waiting atomic.Int64
	done    chan struct{}
}

func (crc *ResourceCoalescingCache[K, T]) Get(ctx context.Context, key K) (result T, err error) {
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
	result, err = crc.fetch()
	res.value = result
	res.err = err
	close(res.done)

	// store the result in the cache if needed
	if err == nil {
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
