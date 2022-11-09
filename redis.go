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

func (r *redisCache) BatchSet(ctx context.Context, keyvalues ...interface{}) error {
	return r.BatchSetCtx(ctx, keyvalues...)
}

func (r *redisCache) BatchSetCtx(ctx context.Context, keyvalues ...interface{}) error {
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
