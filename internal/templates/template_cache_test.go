package templates

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateCacheOperations(t *testing.T) {
	cache := GetTemplateCache()
	cache.Clear()

	tmpl := template.Must(template.New("test").Parse("Hello {{.}}"))
	cache.Set("test", tmpl)

	got, exists := cache.Get("test")
	assert.True(t, exists)
	assert.Same(t, tmpl, got)

	_, exists = cache.Get("bogus")
	assert.False(t, exists)

	cache.Clear()
	_, exists = cache.Get("test")
	assert.False(t, exists)
}

func TestTemplateCacheConcurrency(t *testing.T) {
	cache := GetTemplateCache()
	cache.Clear()

	tmpl := template.Must(template.New("test").Parse("Hello {{.}}"))

	done := make(chan bool)

	for range 10 {
		go func() {
			cache.Set("test", tmpl)
			_, exists := cache.Get("test")
			assert.True(t, exists)

			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}
