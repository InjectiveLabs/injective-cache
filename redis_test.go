//go:build integration

package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultRedisURL = "localhost:6379"

func TestRedisSimpleCache(t *testing.T) {
	ctx := context.Background()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = defaultRedisURL
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisURL})
	_, err := redisClient.Ping(ctx).Result()
	require.NoError(t, err)
	redisClient.FlushAll(ctx)

	ttl := time.Minute
	simpleRedisCache := NewRedisSimpleCache(redisClient, nil, ttl)

	t.Run("Set and Get", func(t *testing.T) {
		key := "key1"
		value := "value1"
		err := simpleRedisCache.Set(ctx, key, value)
		require.NoError(t, err)

		retrievedValue, err := Get[string](ctx, simpleRedisCache, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
	})

	t.Run("SetWithTTL and Get", func(t *testing.T) {
		type valueStruct struct {
			Value string
		}

		key := "key2"
		value := &valueStruct{Value: "value2"}
		err := simpleRedisCache.SetWithTTL(ctx, key, value, time.Millisecond*250)
		require.NoError(t, err)

		retrievedValue, err := Get[*valueStruct](ctx, simpleRedisCache, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
		assert.Equal(t, value.Value, retrievedValue.Value)
		assert.IsType(t, &valueStruct{}, retrievedValue)

		time.Sleep(time.Millisecond * 300)

		retrievedValue, err = Get[*valueStruct](ctx, simpleRedisCache, key)
		require.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		key := "nonexistent"
		err := simpleRedisCache.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Del", func(t *testing.T) {
		key := "key4"
		value := "value4"
		err := simpleRedisCache.Set(ctx, key, value)
		require.NoError(t, err)

		err = simpleRedisCache.Del(ctx, key)
		require.NoError(t, err)

		err = simpleRedisCache.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Clear", func(t *testing.T) {
		err := simpleRedisCache.Set(ctx, "key4", "value4")
		require.NoError(t, err)
		err = simpleRedisCache.Set(ctx, "key5", "value5")
		require.NoError(t, err)

		err = simpleRedisCache.Clear(ctx)
		require.NoError(t, err)

		err = simpleRedisCache.Get(ctx, "key5", nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
		err = simpleRedisCache.Get(ctx, "key6", nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("SetWithTTL expiration", func(t *testing.T) {
		key := "key7"
		value := "value7"
		ttl := time.Millisecond
		err := simpleRedisCache.SetWithTTL(ctx, key, value, ttl)
		require.NoError(t, err)

		time.Sleep(ttl * 2)

		err = simpleRedisCache.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})
}
