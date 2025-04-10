package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMinifyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Non-GET request should skip minification", func(t *testing.T) {
		router := gin.New()
		router.Use(Minify())

		router.POST("/", func(c *gin.Context) {
			c.String(http.StatusOK, "Original content")
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, "Original content", w.Body.String())
		assert.Empty(t, w.Header().Get("X-Cache"))
	})

	t.Run("GET request with HTML content should be minified", func(t *testing.T) {
		GetMinifyCache().Clear()

		router := gin.New()
		router.Use(Minify())

		router.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, `
				<!DOCTYPE html>
				<html>
					<head>
						<title>Test Page</title>
					</head>
					<body>
						<h1>Hello World</h1>
						<p>This is a test page.</p>
					</body>
				</html>
			`)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, "<!doctype html><html><head><title>Test Page</title></head><body><h1>Hello World</h1><p>This is a test page.</body></html>", w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})

	t.Run("GET request with JSON content should be minified", func(t *testing.T) {
		GetMinifyCache().Clear()

		router := gin.New()
		router.Use(Minify())

		router.GET("/api", func(c *gin.Context) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"name": "Test",
				"items": []string{
					"item1",
					"item2",
				},
			})
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/api", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, `{"items":["item1","item2"],"name":"Test"}`, w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})

	t.Run("GET request with unsupported content type should not be minified", func(t *testing.T) {
		GetMinifyCache().Clear()

		router := gin.New()
		router.Use(Minify())

		router.GET("/text", func(c *gin.Context) {
			c.Header("Content-Type", "text/plain")
			c.String(http.StatusOK, "This is plain text content")
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/text", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, "This is plain text content", w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})
}

func TestMinifyWithCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET request with cached content should return cached content", func(t *testing.T) {
		GetMinifyCache().Clear()

		router := gin.New()
		router.Use(Minify())

		router.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, `
				<!DOCTYPE html>
				<html>
					<body>
						<h1>Test Content</h1>
					</body>
				</html>
			`)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		expectedContent := "<!doctype html><html><body><h1>Test Content</h1></body></html>"
		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))

		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
	})

	t.Run("GET request with expired cache should minify and cache", func(t *testing.T) {
		GetMinifyCache().Clear()

		router := gin.New()
		router.Use(Minify())

		router.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, `
				<!DOCTYPE html>
				<html>
					<body>
						<h1>New Content</h1>
					</body>
				</html>
			`)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		expectedContent := "<!doctype html><html><body><h1>New Content</h1></body></html>"
		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))

		GetMinifyCache().Clear()

		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})

	t.Run("GET request in development mode should not use cache", func(t *testing.T) {
		GetMinifyCache().Clear()

		gin.SetMode(gin.DebugMode)
		defer gin.SetMode(gin.TestMode)

		router := gin.New()
		router.Use(Minify())

		router.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, `
				<!DOCTYPE html>
				<html>
					<body>
						<h1>Dev Content</h1>
					</body>
				</html>
			`)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		expectedContent := "<!doctype html><html><body><h1>Dev Content</h1></body></html>"
		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))

		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})
}
