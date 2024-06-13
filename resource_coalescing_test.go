package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shaj13/libcache"
	"github.com/stretchr/testify/require"
)

func TestResourceCoalescingCache(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	t.Run("coalesce", func(t *testing.T) {
		cache := NewTypedLibCache[string, int](libcache.LRU.New(10), time.Minute)
		rcc := NewResourceCoalescingCache[string, int](cache)

		var wg sync.WaitGroup
		var executed int
		waitToExecute := make(chan struct{})
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(t *testing.T) {
				defer wg.Done()

				fetch := func() (int, error) {
					// wait until all goroutines are blocked
					<-waitToExecute
					executed++
					return 42, nil
				}
				res, err := rcc.Get(ctx, "key", fetch)

				// all goroutines should return the same value
				require.NoError(t, err)
				require.Equal(t, 42, res)
			}(t)
		}

		for {
			<-time.NewTimer(10 * time.Millisecond).C
			if rcc.inFlight["key"].waiting.Load() == 9 {
				break
			}
		}

		// execute the function and wait for all promises
		close(waitToExecute)
		wg.Wait()

		require.Equal(t, 1, executed, "only one goroutine should execute the function")
		require.Nil(t, rcc.inFlight["key"], "promise should be removed from inFlight map")

		t.Run("coalesce with cache", func(t *testing.T) {
			var wg sync.WaitGroup
			var executed int
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(t *testing.T) {
					defer wg.Done()

					fetch := func() (int, error) {
						executed++
						return 0, fmt.Errorf("should not be called")
					}
					res, err := rcc.Get(ctx, "key", fetch)

					require.NoError(t, err, "should not return an error")
					require.Equal(t, 42, res, "should return the value from the first call")
				}(t)
			}

			wg.Wait()
		})
	})

	t.Run("coalesce with context error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)

		cache := NewTypedLibCache[string, int](libcache.LRU.New(10), time.Minute)
		rcc := NewResourceCoalescingCache[string, int](cache)

		var wg sync.WaitGroup
		var executed int
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(t *testing.T) {
				defer wg.Done()

				fetch := func() (int, error) {
					<-ctx.Done()
					executed++
					return 0, ctx.Err()
				}
				_, err := rcc.Get(ctx, "key", fetch)

				require.ErrorIs(t, err, context.Canceled)
			}(t)
		}

		for {
			<-time.NewTimer(10 * time.Millisecond).C
			if rcc.inFlight["key"].waiting.Load() == 9 {
				break
			}
		}

		cancel() // cancel the context
		wg.Wait()
		require.Equal(t, 1, executed, "only one goroutine should execute the function")
	})

	t.Run("set cache error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		cache := NewMockTTLCache(ctrl)

		var gotErr error

		rcc := NewResourceCoalescingCache[string, int](cache)
		rcc.OnErr = func(err error) {
			gotErr = err
		}

		cache.EXPECT().Get(ctx, "key", gomock.Any()).Return(ErrCacheMiss)
		cache.EXPECT().Set(ctx, "key", 42).Return(ErrInvalidValue)

		fetch := func() (int, error) {
			return 42, nil
		}
		res, err := rcc.Get(ctx, "key", fetch)

		require.NoError(t, err, "should not return an error")
		require.Equal(t, 42, res, "should return the value from the function")
		require.ErrorIs(t, gotErr, ErrInvalidValue, "should return the error when storing the cache")
	})
}
