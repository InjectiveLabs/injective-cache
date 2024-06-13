package cache

import (
	"context"
	"testing"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRULibCacheString(t *testing.T) {
	ctx := context.Background()
	stringCache := NewTypedLibCache[string, string](libcache.LRU.New(10), time.Minute)

	t.Run("Set and Get", func(t *testing.T) {
		key := "key1"
		value := "value1"
		err := stringCache.Set(ctx, key, value)
		require.NoError(t, err)

		retrievedValue, err := Get[string](ctx, stringCache, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrievedValue)
	})

	t.Run("Set an invalid value", func(t *testing.T) {
		key := "key2"
		value := 2
		err := stringCache.Set(ctx, key, value)
		require.ErrorIs(t, err, ErrInvalidValue)
	})

	t.Run("Get an invalid value", func(t *testing.T) {
		key := "key3"
		value := "value3"
		err := stringCache.Set(ctx, key, value)
		require.NoError(t, err)

		var retrievedValue int
		err = stringCache.Get(ctx, key, &retrievedValue)
		require.ErrorIs(t, err, ErrInvalidValue)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		key := "nonexistent"
		err := stringCache.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Del", func(t *testing.T) {
		key := "key4"
		value := "value4"
		err := stringCache.Set(ctx, key, value)
		require.NoError(t, err)

		err = stringCache.Del(ctx, key)
		require.NoError(t, err)

		err = stringCache.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Clear", func(t *testing.T) {
		err := stringCache.Set(ctx, "key5", "value5")
		require.NoError(t, err)
		err = stringCache.Set(ctx, "key6", "value6")
		require.NoError(t, err)

		err = stringCache.Clear(ctx)
		require.NoError(t, err)

		err = stringCache.Get(ctx, "key5", nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
		err = stringCache.Get(ctx, "key6", nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("SetWithTTL expiration", func(t *testing.T) {
		key := "key7"
		value := "value7"
		ttl := time.Millisecond
		err := stringCache.SetWithTTL(ctx, key, value, ttl)
		require.NoError(t, err)

		time.Sleep(ttl * 2)

		err = stringCache.Get(ctx, key, nil)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})
}

func TestLRULibCacheStruct(t *testing.T) {
	ctx := context.Background()

	type valueStruct struct {
		Value string
	}

	structCache := NewTypedLibCache[string, *valueStruct](libcache.LRU.New(10), time.Minute)

	t.Run("Set and Get", func(t *testing.T) {
		key := "key1"
		value := &valueStruct{
			Value: "value1",
		}
		err := structCache.Set(ctx, key, value)
		require.NoError(t, err)

		retrievedValue, err := Get[*valueStruct](ctx, structCache, key)
		require.NoError(t, err)
		assert.EqualValues(t, value, retrievedValue)
		assert.Equal(t, value.Value, retrievedValue.Value)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		key := "key2"
		value := valueStruct{
			Value: "value2",
		}
		// setting a non-pointer value
		err := structCache.Set(ctx, key, value)
		require.ErrorIs(t, err, ErrInvalidValue)
	})
}
