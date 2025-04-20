package templates

import (
	"html/template"
	"sync"
	"time"

	"github.com/Dobefu/go-web-starter/internal/cache"
)

const TemplateCacheDuration = 24 * time.Hour

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
	tc.cache.Set(key, tmpl, TemplateCacheDuration)
}

func (tc *TemplateCache) Clear() {
	tc.cache.Clear()
}
