package cache

import (
	"sync"
	"time"
)

type Key interface {
	CacheKey() string
}

type Fetcher[T any] interface {
	Fetch(key Key) (T, error)
}

type CacheManager[T any] interface {
	Get(key Key) (T, error)
}

type CacheManagerImpl[T any] struct {
	mutex   sync.Mutex
	cache   map[string]T
	fetcher Fetcher[T]
}

func NewManager[T any](fetcher Fetcher[T]) CacheManager[T] {
	return &CacheManagerImpl[T]{
		cache:   make(map[string]T),
		fetcher: fetcher,
	}
}

func (c *CacheManagerImpl[T]) Get(key Key) (T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cacheKey := key.CacheKey()
	if val, exists := c.cache[cacheKey]; exists {
		return val, nil
	}

	fetched, err := c.fetcher.Fetch(key)
	if err == nil {
		c.cache[cacheKey] = fetched
		// auto-expire after one hour
		go func() {
			time.Sleep(1 * time.Hour)
			delete(c.cache, cacheKey)
		}()
	}
	return fetched, err
}
