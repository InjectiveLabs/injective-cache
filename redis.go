package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	rediscache "github.com/go-redis/redis/v8"
)

var (
	_ Cache = (*redisCache)(nil)
)

type redisCache struct {
	client *rediscache.Client
	ttl    time.Duration
}

func NewRedisCacheWithClient(ctx context.Context, client *rediscache.Client, ttl time.Duration) *redisCache {
	return &redisCache{
		client: client,
		ttl:    ttl,
	}
}

func NewRedisCache(ctx context.Context, cacheURL string, ttl time.Duration) (*redisCache, error) {
	// only 1 cache for now, no need for ring
	client := rediscache.NewClient(&rediscache.Options{
		Addr: cacheURL,
	})

	// try connect
	pingRes := client.Ping(ctx)
	if err := pingRes.Err(); err != nil {
		return nil, fmt.Errorf("redis cache err: %w", err)
	}

	c := NewRedisCacheWithClient(ctx, client, ttl)

	return c, nil
}

func (r *redisCache) Set(ctx context.Context, key string, value []byte) error {
	return r.SetCtx(ctx, key, value)
}

func (r *redisCache) SetCtx(ctx context.Context, key string, value []byte) error {
	status := r.client.Set(ctx, key, value, r.ttl)
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	return r.GetCtx(ctx, key)
}

func (r *redisCache) GetCtx(ctx context.Context, key string) ([]byte, error) {
	result := r.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == rediscache.Nil {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	return result.Bytes()
}

func (r *redisCache) Del(ctx context.Context, key string) error {
	return r.DelCtx(ctx, key)
}

func (r *redisCache) DelCtx(ctx context.Context, key string) error {
	status := r.client.Del(ctx, key)
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) BatchGet(ctx context.Context, keys ...string) (cachedValues [][]byte, err error) {
	return r.BatchGetCtx(ctx, keys...)
}

func (r *redisCache) BatchGetCtx(ctx context.Context, keys ...string) (cachedValues [][]byte, err error) {
	slice := r.client.MGet(ctx, keys...)
	if err := slice.Err(); err != nil {
		return nil, err
	}

	results, err := slice.Result()
	if err != nil {
		return nil, err
	}

	// we should only store bytes atm
	for _, c := range results {
		if c == nil {
			// value nil should also append to maintain order
			cachedValues = append(cachedValues, nil)
			continue
		}

		if str, isString := c.(string); isString {
			cachedValues = append(cachedValues, []byte(str))
		} else {
			cachedValues = append(cachedValues, c.([]byte))
		}
	}
	return cachedValues, nil
}

func (r *redisCache) BatchSet(ctx context.Context, keyvalues ...any) error {
	return r.BatchSetCtx(ctx, keyvalues...)
}

func (r *redisCache) BatchSetCtx(ctx context.Context, keyvalues ...any) error {
	if len(keyvalues)%2 != 0 {
		return errors.New("keyvalues len must be even")
	}

	// since atm redis MSET does not support set value with ttl
	// => use TxPipeline instead, it's atomic on redis server
	// however the pipeline should be short to make sure it does not block server so long
	pipeline := r.client.TxPipeline()
	for i := 0; i < len(keyvalues); i += 2 {
		key, ok := keyvalues[i].(string)
		if !ok {
			return fmt.Errorf("%w at index %d: expected string, got %T", ErrInvalidKey, i, keyvalues[i])
		}

		value, ok := keyvalues[i+1].([]byte)
		if !ok {
			return fmt.Errorf("%w at index %d: expected []byte, got %T", ErrInvalidValue, i, keyvalues[i])
		}

		pipeline.Set(ctx, key, value, r.ttl)
	}

	// exec the command
	statuses, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	for _, s := range statuses {
		if err = s.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

func (r *redisCache) CloseCtx(_ context.Context) error {
	return r.Close()
}

func (r *redisCache) IsRunning(ctx context.Context) bool {
	if r.client == nil {
		return false
	}
	return r.client.Ping(ctx).Err() == nil
}

var _ TTLCache = (*RedisSimpleCache)(nil)

// RedisSimpleCache is a redis cache implementation using go-redis/v8
type RedisSimpleCache struct {
	// client is the redis client
	client *rediscache.Client
	// ttl is the default time-to-live for cache entries: 0 means no expiration
	ttl time.Duration
	// codec allows to specify a custom codec for encoding/decoding values
	// json is used by default
	codec Codec
}

// NewRedisSimpleCache creates a new RedisSimpleCache instance
func NewRedisSimpleCache(client *rediscache.Client, codec Codec, ttl time.Duration) *RedisSimpleCache {
	if codec == nil {
		codec = &JsonCodec{}
	}
	return &RedisSimpleCache{
		client: client,
		ttl:    ttl,
		codec:  codec,
	}
}

func (r *RedisSimpleCache) Set(ctx context.Context, key any, value any) (err error) {
	return r.SetWithTTL(ctx, key, value, r.ttl)
}

func (r *RedisSimpleCache) SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) (err error) {
	k, err := keyToString(key)
	if err != nil {
		return err
	}
	data, err := r.codec.Encode(value)
	if err != nil {
		return ErrInvalidValue
	}
	status := r.client.Set(ctx, k, data, ttl)
	if err = status.Err(); err != nil {
		return fmt.Errorf("setting key %s: %w", k, err)
	}
	return nil
}

func (r *RedisSimpleCache) Get(ctx context.Context, key any, value any) (err error) {
	k, err := keyToString(key)
	if err != nil {
		return err
	}
	data, err := r.client.Get(ctx, k).Bytes()
	if errors.Is(err, rediscache.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("getting key %s: %w", k, err)
	}

	err = r.codec.Decode(data, value)
	if err != nil {
		return fmt.Errorf("decoding %s value: %w", k, err)
	}
	return nil
}

func (r *RedisSimpleCache) Del(ctx context.Context, keys ...any) (err error) {
	ks, err := keysToString(keys...)
	if err != nil {
		return err
	}
	status := r.client.Del(ctx, ks...)
	if err = status.Err(); err != nil {
		return fmt.Errorf("deleting keys: %w", err)
	}
	return nil
}

func (r *RedisSimpleCache) Clear(ctx context.Context) (err error) {
	status := r.client.FlushDB(ctx)
	if err = status.Err(); err != nil {
		return fmt.Errorf("clearing cache: %w", err)
	}
	return nil
}

// keyToString converts a key to a string
func keyToString(k any) (string, error) {
	if s, ok := k.(string); ok {
		return s, nil
	}
	s := fmt.Sprintf("%v", k)
	if s == "" {
		return "", ErrInvalidKey
	}
	return s, nil
}

// keys converts a list of keys to a list of strings
func keysToString(keys ...any) ([]string, error) {
	ks := make([]string, 0, len(keys))
	for _, k := range keys {
		valid, err := keyToString(k)
		if err != nil {
			return nil, err
		}
		ks = append(ks, valid)
	}
	return ks, nil
}
