package cache

import (
	"sync"
	"time"
)

type Cache[T any] struct {
	content map[string]cacheEntry[T]
	mutex   sync.RWMutex
}

type cacheEntry[T any] struct {
	content    T
	expiration time.Time
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		content: make(map[string]cacheEntry[T]),
	}
}

func (c *Cache[T]) Get(key string) (zero T, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if entry, exists := c.content[key]; exists {
		if time.Now().Before(entry.expiration) {
			return entry.content, true
		}

		delete(c.content, key)
	}

	return zero, false
}

func (c *Cache[T]) Set(key string, content T, expiration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.content[key] = cacheEntry[T]{
		content:    content,
		expiration: time.Now().Add(expiration),
	}
}

func (c *Cache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.content = make(map[string]cacheEntry[T])
}

func (c *Cache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.content, key)
}

func (c *Cache[T]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.content)
}
