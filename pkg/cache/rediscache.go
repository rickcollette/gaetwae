package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{client: rdb, ctx: context.Background()}
}

func (c *RedisCache) Set(key string, value []byte, expirationTime int) error {
	return c.client.Set(c.ctx, key, value, time.Duration(expirationTime)*time.Second).Err()
}

func (c *RedisCache) Get(key string) ([]byte, error) {
	val, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}
