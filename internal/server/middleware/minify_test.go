package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MinifyMockRedis struct {
	mock.Mock
}

func (m *MinifyMockRedis) Close() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MinifyMockRedis) Get(ctx context.Context, key string) (*redisClient.StringCmd, error) {
	args := m.Called(ctx, key)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StringCmd), args.Error(1)
}

func (m *MinifyMockRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) (*redisClient.StatusCmd, error) {
	args := m.Called(ctx, key, value, expiration)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StatusCmd), args.Error(1)
}

func (m *MinifyMockRedis) GetRange(ctx context.Context, key string, start, end int64) (*redisClient.StringCmd, error) {
	args := m.Called(ctx, key, start, end)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StringCmd), args.Error(1)
}

func (m *MinifyMockRedis) SetRange(ctx context.Context, key string, offset int64, value string) (*redisClient.IntCmd, error) {
	args := m.Called(ctx, key, offset, value)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.IntCmd), args.Error(1)
}

func (m *MinifyMockRedis) FlushDB(ctx context.Context) (*redisClient.StatusCmd, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StatusCmd), args.Error(1)
}

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

func TestMinifyWithRedisCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET request with cached content should return cached content", func(t *testing.T) {
		mockRedis := new(MinifyMockRedis)

		mockStringCmd := &redisClient.StringCmd{}
		mockStringCmd.SetVal("<html><body>Minified cached content</body></html>")

		mockRedis.On("Get", mock.Anything, "minify:GET:/").Return(mockStringCmd, nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("redis", mockRedis)
			c.Next()
		})
		router.Use(Minify())

		handlerCalled := false
		router.GET("/", func(c *gin.Context) {
			handlerCalled = true
			c.String(http.StatusOK, "This should not be returned")
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, "<html><body>Minified cached content</body></html>", w.Body.String())
		assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
		assert.False(t, handlerCalled, "Handler should not be called when cached content is available")
		mockRedis.AssertExpectations(t)
	})

	t.Run("GET request with no cached content should minify and cache", func(t *testing.T) {
		mockRedis := new(MinifyMockRedis)

		mockRedis.On("Get", mock.Anything, "minify:GET:/").Return(nil, redisClient.Nil)

		mockStatusCmd := &redisClient.StatusCmd{}
		mockStatusCmd.SetVal("OK")
		mockRedis.On("Set", mock.Anything, "minify:GET:/", mock.Anything, time.Hour).Return(mockStatusCmd, nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("redis", mockRedis)
			c.Next()
		})
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

		assert.Equal(t, "<!doctype html><html><body><h1>Test Content</h1></body></html>", w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
		mockRedis.AssertExpectations(t)
	})

	t.Run("GET request with Redis error should still minify content", func(t *testing.T) {
		mockRedis := new(MinifyMockRedis)

		mockRedis.On("Get", mock.Anything, "minify:GET:/").Return(nil, assert.AnError)

		mockStatusCmd := &redisClient.StatusCmd{}
		mockStatusCmd.SetVal("OK")
		mockRedis.On("Set", mock.Anything, "minify:GET:/", mock.Anything, time.Hour).Return(mockStatusCmd, nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("redis", mockRedis)
			c.Next()
		})
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

		assert.Equal(t, "<!doctype html><html><body><h1>Test Content</h1></body></html>", w.Body.String())
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
		mockRedis.AssertExpectations(t)
	})
}
