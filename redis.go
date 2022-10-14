package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	rediscache "github.com/go-redis/redis/v8"
)

type redisCache struct {
	ctx    context.Context
	client *rediscache.Client
	ttl    time.Duration
	Cache
}

func NewRedisCache(ctx context.Context, cacheURL string, ttl time.Duration) (Cache, error) {
	// only 1 cache for now, no need for ring
	c := rediscache.NewClient(&rediscache.Options{
		Addr: cacheURL,
	})

	// try connect
	pingRes := c.Ping(ctx)
	if err := pingRes.Err(); err != nil {
		return nil, fmt.Errorf("redis cache err: %w", err)
	}

	return &redisCache{
		ctx:    ctx,
		client: c,
		ttl:    ttl,
	}, nil
}

func (r *redisCache) Set(key string, value []byte) error {
	status := r.client.Set(r.ctx, key, value, r.ttl)
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Get(key string) ([]byte, error) {
	result := r.client.Get(r.ctx, key)
	if err := result.Err(); err != nil {
		return nil, err
	}
	return result.Bytes()
}

func (r *redisCache) Del(key string) error {
	status := r.client.Del(r.ctx, key)
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) BatchGet(keys ...string) (cachedValues [][]byte, err error) {
	slice := r.client.MGet(r.ctx, keys...)
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

func (r *redisCache) BatchSet(keyvalues ...interface{}) error {
	if len(keyvalues)%2 != 0 {
		return errors.New("keyvalues len must be even")
	}

	// since atm redis MSET does not support set value with ttl
	// => use TxPipeline instead, it's atomic on redis server
	// however the pipeline should be short to make sure it does not block server so long
	pipeline := r.client.TxPipeline()
	for i := 0; i < len(keyvalues); i += 2 {
		key := keyvalues[i].(string)
		value := keyvalues[i+1].([]byte)
		pipeline.Set(r.ctx, key, value, r.ttl)
	}

	// exec the command
	statuses, err := pipeline.Exec(r.ctx)
	if err != nil {
		return err
	}

	for _, s := range statuses {
		if err := s.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *redisCache) Close() error {
	return r.client.Close()
}
