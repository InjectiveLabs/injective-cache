//go:build integration

package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

var defaultRedisURL = "localhost:6379"

func TestRedisCacheV2(t *testing.T) {
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
	redisCacheV2 := NewRedisCacheV2(redisClient, nil, ttl)

	t.Run("Set and Get", func(t *testing.T) {
		key := "key1"
		value := "value1"
		err := redisCacheV2.Set(ctx, key, value)
		require.NoError(t, err)

		retrievedValue, err := Get[string](ctx, redisCacheV2, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
	})

	t.Run("SetWithTTL and Get", func(t *testing.T) {
		type valueStruct struct {
			Value string
		}

		key := "key2"
		value := &valueStruct{Value: "value2"}
		err := redisCacheV2.SetWithTTL(ctx, key, value, time.Millisecond*250)
		require.NoError(t, err)

		retrievedValue, err := Get[*valueStruct](ctx, redisCacheV2, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
		assert.Equal(t, value.Value, retrievedValue.Value)
		assert.IsType(t, &valueStruct{}, retrievedValue)

		time.Sleep(time.Millisecond * 300)

		retrievedValue, err = Get[*valueStruct](ctx, redisCacheV2, key)
		require.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		key := "nonexistent"
		err := redisCacheV2.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Del", func(t *testing.T) {
		key := "key4"
		value := "value4"
		err := redisCacheV2.Set(ctx, key, value)
		require.NoError(t, err)

		err = redisCacheV2.Del(ctx, key)
		require.NoError(t, err)

		err = redisCacheV2.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Clear", func(t *testing.T) {
		err := redisCacheV2.Set(ctx, "key4", "value4")
		require.NoError(t, err)
		err = redisCacheV2.Set(ctx, "key5", "value5")
		require.NoError(t, err)

		err = redisCacheV2.Clear(ctx)
		require.NoError(t, err)

		err = redisCacheV2.Get(ctx, "key5", nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
		err = redisCacheV2.Get(ctx, "key6", nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("SetWithTTL expiration", func(t *testing.T) {
		key := "key7"
		value := "value7"
		ttl := time.Millisecond
		err := redisCacheV2.SetWithTTL(ctx, key, value, ttl)
		require.NoError(t, err)

		time.Sleep(ttl * 2)

		err = redisCacheV2.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})
}
