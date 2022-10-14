package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	cache, err := NewRedisCache(context.Background(), "localhost:6379", 5*time.Second)
	assert.Nil(t, err)

	cache.Set("hello", []byte("word"))
	v, err := cache.Get("hello")
	assert.Nil(t, err)

	assert.Equal(t, []byte("word"), v)
}
