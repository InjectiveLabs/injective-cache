package cache

import (
	"context"
	"sync"
	"sync/atomic"
)

type Promise[T any] struct {
	done  <-chan struct{}
	value T
	err   error
}

func NewPromise[T any]() (*Promise[T], func(T, error)) {
	ch := make(chan struct{})
	p := &Promise[T]{
		done: ch,
	}
	resolve := func(value T, err error) {
		p.value = value
		p.err = err
		close(ch)
	}
	return p, resolve
}

type inFlightPromise[T any] struct {
	*Promise[T]
	waiting atomic.Int64
}

type CachedResourceCoalescing[K comparable, T any] struct {
	OnErr    func(error)
	cache    TTLCache
	cacheMX  sync.RWMutex
	inFlight map[K]*inFlightPromise[T]
	once     sync.Once
}

func NewCachedResourceCoalescing[K comparable, T any](cache TTLCache) *CachedResourceCoalescing[K, T] {
	return &CachedResourceCoalescing[K, T]{
		cache:    cache,
		inFlight: make(map[K]*inFlightPromise[T]),
	}
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
	promise, found := crc.inFlight[key]
	if found {
		promise.waiting.Add(1)
		defer promise.waiting.Add(-1)
		crc.cacheMX.Unlock()
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-promise.done:
			return promise.value, promise.err
		}
	}

	// not found, create a new promise and query the result
	var resolve func(T, error)
	promise = new(inFlightPromise[T])
	promise.Promise, resolve = NewPromise[T]()
	crc.inFlight[key] = promise
	crc.cacheMX.Unlock()

	// execute the function
	v, store, err := get()

	// resolve the promise to return the results
	resolve(v, err)

	// store the result in the cache if needed
	if store {
		if setErr := crc.cache.Set(ctx, key, v); setErr != nil && crc.OnErr != nil {
			crc.OnErr(setErr)
		}
	}

	// remove the promise from the in-flight map
	crc.cacheMX.Lock()
	delete(crc.inFlight, key)
	crc.cacheMX.Unlock()

	return v, err
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
