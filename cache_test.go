package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGet(t *testing.T) {
	cache, err := NewLibcache(0, 5*time.Second)
	assert.Nil(t, err)

	cache.Set("key1", []byte("hello1"))
	cache.Set("key2", []byte("hello2"))

	v, err := cache.Get("key2")
	assert.Nil(t, err)

	assert.Equal(t, []byte("hello2"), v)
}

func TestBatchSetGet(t *testing.T) {
	cache, err := NewLibcache(0, 5*time.Second)
	assert.Nil(t, err)

	cache.BatchSet("key1", []byte("value1"), "key2", []byte("value2"), "key3", []byte("value3"))

	v, err := cache.BatchGet("key3", "key2")
	assert.Nil(t, err)

	assert.Equal(t, []byte("value3"), v[0])
	assert.Equal(t, []byte("value2"), v[1])
}

func TestRedisSet(t *testing.T) {
	ctx := context.Background()
	cache, err := NewRedisCache(ctx, "localhost:6379", 5*time.Second)
	assert.Nil(t, err)

	cache.SetCtx(ctx, "hello", []byte("word"))
	v, err := cache.GetCtx(ctx, "hello")
	assert.Nil(t, err)

	assert.Equal(t, []byte("word"), v)

	t.Run("batch set", func(t *testing.T) {
		require.Nil(t, cache.BatchSetCtx(ctx, "hello", []byte("world")))
		v, err := cache.GetCtx(ctx, "hello")
		assert.Nil(t, err)
		assert.Equal(t, []byte("world"), v)

		require.ErrorIs(t, cache.BatchSetCtx(ctx, true, []byte("world")), ErrInvalidKey)
		require.ErrorIs(t, cache.BatchSetCtx(ctx, "hello", false), ErrInvalidValue)
	})
}
