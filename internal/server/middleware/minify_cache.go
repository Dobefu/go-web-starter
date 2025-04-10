package middleware

import (
	"sync"
	"time"

	"github.com/Dobefu/go-web-starter/internal/cache"
)

type MinifyCache struct {
	cache *cache.Cache[[]byte]
}

var (
	minifyCache *MinifyCache
	once        sync.Once
)

func GetMinifyCache() *MinifyCache {
	once.Do(func() {
		minifyCache = &MinifyCache{
			cache: cache.New[[]byte](),
		}
	})

	return minifyCache
}

func (mc *MinifyCache) Get(key string) []byte {
	content, exists := mc.cache.Get(key)

	if !exists {
		return nil
	}

	return content
}

func (mc *MinifyCache) Set(key string, content []byte, expiration time.Duration) {
	mc.cache.Set(key, content, expiration)
}

func (mc *MinifyCache) Clear() {
	mc.cache.Clear()
}
