package cache

import (

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedCache struct {
	client *memcache.Client
}

func NewMemcachedCache(server ...string) *MemcachedCache {
	mc := memcache.New(server...)
	return &MemcachedCache{client: mc}
}

func (c *MemcachedCache) Set(key string, value []byte, expirationTime int) error {
	item := &memcache.Item{Key: key, Value: value, Expiration: int32(expirationTime)}
	return c.client.Set(item)
}

func (c *MemcachedCache) Get(key string) ([]byte, error) {
	item, err := c.client.Get(key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}
