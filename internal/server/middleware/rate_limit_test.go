package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-gonic/gin"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testClientIP = "127.0.0.1"
	redisNilErr  = "redis: nil"
)

type MockRedis struct {
	mock.Mock
	tokens   int
	lastTime time.Time
}

func (m *MockRedis) Close() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockRedis) Get(ctx context.Context, key string) (*redisClient.StringCmd, error) {
	args := m.Called(ctx, key)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StringCmd), args.Error(1)
}

func (m *MockRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) (*redisClient.StatusCmd, error) {
	args := m.Called(ctx, key, value, expiration)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StatusCmd), args.Error(1)
}

func (m *MockRedis) GetRange(ctx context.Context, key string, start, end int64) (*redisClient.StringCmd, error) {
	args := m.Called(ctx, key, start, end)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StringCmd), args.Error(1)
}

func (m *MockRedis) SetRange(ctx context.Context, key string, offset int64, value string) (*redisClient.IntCmd, error) {
	args := m.Called(ctx, key, offset, value)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.IntCmd), args.Error(1)
}

func setupMockRedis(tokens int, redisErr error, now time.Time) (*MockRedis, *RateLimiter) {
	mockRedis := new(MockRedis)
	mockRedis.tokens = tokens
	mockRedis.lastTime = now

	var mockCmd *redisClient.StringCmd

	if redisErr != nil {
		mockCmd = createMockStringCmd("", redisErr)
	} else {
		mockCmd = createMockStringCmd(fmt.Sprintf("%d:%d", tokens, now.Unix()), nil)
	}

	mockRedis.On("Get", mock.Anything, mock.Anything).Return(mockCmd, redisErr)

	if redisErr == nil || redisErr.Error() == redisNilErr {
		mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			createMockStatusCmd(nil),
			nil,
		)
	}

	limiter := NewRateLimiterWithRedis(mockRedis, 5, time.Second)
	limiter.timeNow = func() time.Time { return now }

	return mockRedis, limiter
}

func setupTestRouter(limiter *RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		if !limiter.Allow(getClientIP(c)) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()

			return
		}

		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func createMockStringCmd(val string, err error) *redisClient.StringCmd {
	cmd := redisClient.NewStringCmd(context.Background())

	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(val)
	}

	return cmd
}

func createMockStatusCmd(err error) *redisClient.StatusCmd {
	cmd := redisClient.NewStatusCmd(context.Background())

	if err != nil {
		cmd.SetErr(err)
	}

	return cmd
}

func TestRateLimiterAllow(t *testing.T) {
	tests := []struct {
		name          string
		clientID      string
		tokens        int
		redisError    error
		expectedAllow bool
	}{
		{
			name:          "new client",
			clientID:      testClientIP,
			redisError:    errors.New(redisNilErr),
			expectedAllow: true,
		},
		{
			name:          "existing client with tokens",
			clientID:      testClientIP,
			tokens:        4,
			expectedAllow: true,
		},
		{
			name:          "no tokens left",
			clientID:      testClientIP,
			tokens:        0,
			expectedAllow: false,
		},
		{
			name:          "redis error",
			clientID:      testClientIP,
			redisError:    errors.New("connection error"),
			expectedAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis, limiter := setupMockRedis(tt.tokens, tt.redisError, time.Now())
			result := limiter.Allow(tt.clientID)
			assert.Equal(t, tt.expectedAllow, result)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		tokens       int
		getError     error
		expectedCode int
		setupRequest func(*http.Request)
	}{
		{
			name:         "successful request",
			tokens:       5,
			expectedCode: http.StatusOK,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = testClientIP + ":1234"
			},
		},
		{
			name:         "rate limited request",
			tokens:       0,
			expectedCode: http.StatusTooManyRequests,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = "rate-limited-ip:1234"
			},
		},
		{
			name:         "redis error",
			tokens:       0,
			getError:     errors.New("redis error"),
			expectedCode: http.StatusOK,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = testClientIP + ":1234"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis, limiter := setupMockRedis(tt.tokens, tt.getError, time.Now())
			router := setupTestRouter(limiter)

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)

			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			expectedIP: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "No headers or remote addr",
			remoteAddr: "",
			expectedIP: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			ip := getClientIP(c)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestNewRateLimiter(t *testing.T) {
	originalNew := redis.New
	defer func() { redis.New = originalNew }()

	tests := []struct {
		name          string
		redisError    error
		expectError   bool
		errorContains string
	}{
		{
			name:          "successful creation",
			redisError:    nil,
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "redis error",
			redisError:    errors.New("redis connection error"),
			expectError:   true,
			errorContains: "failed to create Redis client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis.New = func(cfg config.Redis, log *logger.Logger) (*redis.Redis, error) {
				if tt.redisError != nil {
					return nil, tt.redisError
				}
				return &redis.Redis{}, nil
			}

			limiter, err := NewRateLimiter(5, time.Second)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, limiter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, limiter)
				assert.Equal(t, 5, limiter.capacity)
				assert.Equal(t, time.Second, limiter.rate)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	originalNew := redis.New
	defer func() { redis.New = originalNew }()

	tests := []struct {
		name         string
		redisError   error
		expectedCode int
	}{
		{
			name:         "successful creation",
			redisError:   nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "redis error",
			redisError:   errors.New("redis connection error"),
			expectedCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis.New = func(cfg config.Redis, log *logger.Logger) (*redis.Redis, error) {
				if tt.redisError != nil {
					return nil, tt.redisError
				}
				return &redis.Redis{}, nil
			}

			if tt.expectedCode == 0 {
				assert.Panics(t, func() {
					RateLimit(5, time.Second)
				})
				return
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(RateLimit(5, time.Second))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)
			req.RemoteAddr = testClientIP + ":1234"

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
