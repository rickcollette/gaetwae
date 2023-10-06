package cache

import (
	"fmt"
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

type InMemoryCache struct {
	capacity int
	items    map[string]item
	mu       sync.Mutex
}

func NewInMemoryCache(capacity int) *InMemoryCache {
	return &InMemoryCache{
		capacity: capacity,
		items:    make(map[string]item),
	}
}

func (c *InMemoryCache) Set(key string, value []byte, expirationTime int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.capacity {
		c.evictOldest()
	}

	expiration := time.Now().Add(time.Duration(expirationTime) * time.Second).UnixNano()
	c.items[key] = item{
		value:      value,
		expiration: expiration,
	}
    return nil  // Added this line to return nil error
}

func (c *InMemoryCache) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		return nil, fmt.Errorf("item not found")
	}

	if time.Now().UnixNano() > item.expiration {
		return nil, fmt.Errorf("item expired")
	}

	value, ok := item.value.([]byte)
	if !ok {
		return nil, fmt.Errorf("error converting item value to []byte")
	}

	return value, nil
}

func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestExpiration int64

	for key, item := range c.items {
		if oldestKey == "" || item.expiration < oldestExpiration {
			oldestKey = key
			oldestExpiration = item.expiration
		}
	}

	delete(c.items, oldestKey)
}