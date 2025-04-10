package templates

import (
	"html/template"
	"sync"

	"github.com/Dobefu/go-web-starter/internal/cache"
)

type TemplateCache struct {
	cache *cache.Cache[*template.Template]
}

var (
	templateCache *TemplateCache
	once          sync.Once
)

func GetTemplateCache() *TemplateCache {
	once.Do(func() {
		templateCache = &TemplateCache{
			cache: cache.New[*template.Template](),
		}
	})

	return templateCache
}

func (tc *TemplateCache) Get(key string) (*template.Template, bool) {
	return tc.cache.Get(key)
}

func (tc *TemplateCache) Set(key string, tmpl *template.Template) {
	tc.cache.Set(key, tmpl, 0)
}

func (tc *TemplateCache) Clear() {
	tc.cache.Clear()
}
