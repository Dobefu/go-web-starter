package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CacheHeaders())

	t.Run("Static assets get cached for 7 days", func(t *testing.T) {
		router.GET("/static/style.css", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/static/style.css", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		headers := w.Header()
		assert.Equal(t, "public, max-age=604800", headers.Get("Cache-Control"))
		assert.NotEmpty(t, headers.Get("Expires"))
	})

	t.Run("Dynamic content gets no-cache", func(t *testing.T) {
		router.GET("/api/data", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/api/data", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		headers := w.Header()
		assert.Equal(t, "no-cache, must-revalidate", headers.Get("Cache-Control"))
		assert.Equal(t, "no-cache", headers.Get("Pragma"))
	})

	t.Run("Non-GET requests skip cache headers", func(t *testing.T) {
		router.POST("/api/data", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/api/data", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		headers := w.Header()
		assert.Empty(t, headers.Get("Cache-Control"))
		assert.Empty(t, headers.Get("Expires"))
		assert.Empty(t, headers.Get("Pragma"))
	})

	t.Run("Short paths are not treated as static assets", func(t *testing.T) {
		router.GET("/a", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/a", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		headers := w.Header()
		assert.Equal(t, "no-cache, must-revalidate", headers.Get("Cache-Control"))
		assert.Equal(t, "no-cache", headers.Get("Pragma"))
	})
}
