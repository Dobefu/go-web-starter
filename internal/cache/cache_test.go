package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	c := New[string]()

	assert.NotNil(t, c)
	assert.NotNil(t, c.content)
}

func TestGetSet(t *testing.T) {
	t.Parallel()

	c := New[string]()

	c.Set("key1", "value1", time.Hour)
	value, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	value, ok = c.Get("bogus")
	assert.False(t, ok)
	assert.Empty(t, value)
}

func TestExpiration(t *testing.T) {
	t.Parallel()

	c := New[string]()

	c.Set("key1", "value1", (100 * time.Millisecond))

	value, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	time.Sleep(150 * time.Millisecond)

	value, ok = c.Get("key1")
	assert.False(t, ok)
	assert.Empty(t, value)
}

func TestClear(t *testing.T) {
	t.Parallel()

	c := New[string]()

	c.Set("key1", "value1", time.Hour)
	c.Set("key2", "value2", time.Hour)

	c.Clear()

	assert.Equal(t, 0, c.Size())

	_, ok := c.Get("key1")
	assert.False(t, ok)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	c := New[string]()

	c.Set("key1", "value1", time.Hour)

	c.Delete("key1")

	_, ok := c.Get("key1")
	assert.False(t, ok)

	assert.NotPanics(t, func() { c.Delete("bogus") })
}
