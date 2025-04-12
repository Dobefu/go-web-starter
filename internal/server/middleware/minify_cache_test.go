package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMinifyCache(t *testing.T) {
	cache1 := GetMinifyCache()
	assert.NotNil(t, cache1)
	assert.NotNil(t, cache1.cache)

	cache2 := GetMinifyCache()
	assert.Equal(t, cache1, cache2)
}

func TestMinifyCache_GetSet(t *testing.T) {
	cache := GetMinifyCache()
	key := "test-key"
	content := []byte("test content")

	result := cache.Get(key)
	assert.Nil(t, result)

	cache.Set(key, content, time.Hour)
	result = cache.Get(key)
	assert.Equal(t, content, result)
}

func TestMinifyCache_Expiration(t *testing.T) {
	cache := GetMinifyCache()
	key := "test-key"
	content := []byte("test content")

	cache.Set(key, content, 10*time.Millisecond)

	result := cache.Get(key)
	assert.Equal(t, content, result)

	time.Sleep(15 * time.Millisecond)

	result = cache.Get(key)
	assert.Nil(t, result)
}

func TestMinifyCache_Clear(t *testing.T) {
	cache := GetMinifyCache()
	key1 := "key1"
	key2 := "key2"

	content1 := []byte("content1")
	content2 := []byte("content2")

	cache.Set(key1, content1, time.Hour)
	cache.Set(key2, content2, time.Hour)

	assert.Equal(t, content1, cache.Get(key1))
	assert.Equal(t, content2, cache.Get(key2))

	cache.Clear()

	assert.Nil(t, cache.Get(key1))
	assert.Nil(t, cache.Get(key2))
}
