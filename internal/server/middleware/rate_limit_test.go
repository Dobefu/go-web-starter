package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

func (m *MockRedis) FlushDB(ctx context.Context) (*redisClient.StatusCmd, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StatusCmd), args.Error(1)
}

func (m *MockRedis) SetWithTTL(ctx context.Context, key string, value any) (*redisClient.StatusCmd, error) {
	args := m.Called(ctx, key, value)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*redisClient.StatusCmd), args.Error(1)
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

func setupTestRouter(handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("test-session", store))
	router.Use(handler)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestRateLimiterAllow(t *testing.T) {
	tests := []struct {
		name          string
		clientID      string
		tokens        int
		redisError    error
		setError      error
		malformedData bool
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
		{
			name:          "malformed data",
			clientID:      testClientIP,
			malformedData: true,
			expectedAllow: true,
		},
		{
			name:          "set error on no tokens",
			clientID:      testClientIP,
			tokens:        0,
			setError:      errors.New("set error"),
			expectedAllow: false,
		},
		{
			name:          "set error on update",
			clientID:      testClientIP,
			tokens:        5,
			setError:      errors.New("set error"),
			expectedAllow: true,
		},
		{
			name:          "token refill",
			clientID:      testClientIP,
			tokens:        0,
			expectedAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			mockRedis := new(MockRedis)

			var mockCmd *redisClient.StringCmd

			if tt.malformedData {
				mockCmd = createMockStringCmd("invalid:data", nil)
				mockRedis.On("Get", mock.Anything, mock.Anything).Return(mockCmd, nil)
			} else if tt.redisError != nil {
				if tt.redisError.Error() == redisNilErr {
					mockCmd = createMockStringCmd("", tt.redisError)
					mockRedis.On("Get", mock.Anything, mock.Anything).Return(mockCmd, tt.redisError)
					mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
						createMockStatusCmd(nil),
						nil,
					)
				} else {
					mockCmd = createMockStringCmd("", tt.redisError)
					mockRedis.On("Get", mock.Anything, mock.Anything).Return(mockCmd, tt.redisError)
				}
			} else {
				lastUpdate := now

				if tt.name == "token refill" {
					lastUpdate = now.Add(-2 * time.Second)
				}

				mockCmd = createMockStringCmd(fmt.Sprintf("%d:%d", tt.tokens, lastUpdate.Unix()), nil)
				mockRedis.On("Get", mock.Anything, mock.Anything).Return(mockCmd, nil)

				if tt.name == "existing client with tokens" || tt.name == "no tokens left" || tt.name == "set error on no tokens" || tt.name == "set error on update" || tt.name == "token refill" {
					mockRedis.On("SetWithTTL", mock.Anything, mock.Anything, mock.Anything).Return(
						createMockStatusCmd(tt.setError),
						tt.setError,
					)
				} else {
					mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
						createMockStatusCmd(tt.setError),
						tt.setError,
					)
				}
			}

			limiter := NewRateLimiterWithRedis(mockRedis, 5, time.Second)
			limiter.timeNow = func() time.Time { return now }

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
		redisEnabled bool
		expectedCode int
		setupRequest func(*http.Request)
	}{
		{
			name:         "successful request",
			tokens:       5,
			redisEnabled: true,
			expectedCode: http.StatusOK,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = testClientIP + ":1234"
			},
		},
		{
			name:         "rate limited request",
			tokens:       0,
			redisEnabled: true,
			expectedCode: http.StatusTooManyRequests,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = "rate-limited-ip:1234"
			},
		},
		{
			name:         "redis error",
			tokens:       0,
			getError:     errors.New("redis error"),
			redisEnabled: true,
			expectedCode: http.StatusOK,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = testClientIP + ":1234"
			},
		},
		{
			name:         "redis disabled",
			redisEnabled: false,
			expectedCode: http.StatusOK,
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = testClientIP + ":1234"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalConfig := config.DefaultConfig
			defer func() { config.DefaultConfig = originalConfig }()

			config.DefaultConfig.Redis.Enable = tt.redisEnabled

			now := time.Now()
			mockRedis := new(MockRedis)
			var handler gin.HandlerFunc

			if tt.redisEnabled {
				req, _ := http.NewRequest("GET", "/test", nil)
				tt.setupRequest(req)
				clientIP := strings.Split(req.RemoteAddr, ":")[0]
				key := fmt.Sprintf("rate_limit:%s", clientIP)
				value := fmt.Sprintf("%d:%d", tt.tokens, now.Unix())

				mockCmd := createMockStringCmd(value, tt.getError)
				mockRedis.On("Get", mock.Anything, key).Return(mockCmd, tt.getError)

				if tt.getError == nil || tt.getError.Error() == redisNilErr {
					if tt.getError != nil && tt.getError.Error() == redisNilErr {
						mockRedis.On("Set", mock.Anything, key, mock.Anything, mock.Anything).Return(
							createMockStatusCmd(nil),
							nil,
						)
					} else {
						mockRedis.On("SetWithTTL", mock.Anything, key, mock.Anything).Return(
							createMockStatusCmd(nil),
							nil,
						)
					}
				}

				limiter := NewRateLimiterWithRedis(mockRedis, 5, time.Second)
				limiter.timeNow = func() time.Time { return now }

				handler = limiter.Middleware()
			} else {
				handler = RateLimit(5, time.Second)
			}

			router := setupTestRouter(handler)
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)

			tt.setupRequest(req)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.redisEnabled {
				mockRedis.AssertExpectations(t)
			}
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
	tests := []struct {
		name         string
		redisEnabled bool
		tokens       int
		redisError   error
		expectedCode int
	}{
		{
			name:         "redis disabled",
			redisEnabled: false,
			expectedCode: http.StatusOK,
		},
		{
			name:         "redis enabled, request allowed",
			redisEnabled: true,
			tokens:       5,
			expectedCode: http.StatusOK,
		},
		{
			name:         "redis enabled, request rate limited",
			redisEnabled: true,
			tokens:       0,
			expectedCode: http.StatusTooManyRequests,
		},
		{
			name:         "redis error returns 500",
			redisEnabled: true,
			redisError:   errors.New("redis connection error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalConfig := config.DefaultConfig
			defer func() { config.DefaultConfig = originalConfig }()

			config.DefaultConfig.Redis.Enable = tt.redisEnabled

			if tt.redisEnabled {
				if tt.redisError != nil {
					originalNew := redis.New
					defer func() { redis.New = originalNew }()

					redis.New = func(cfg config.Redis, log *logger.Logger) (*redis.Redis, error) {
						return nil, tt.redisError
					}

					handler := RateLimit(5, time.Second)
					router := setupTestRouter(handler)
					w := httptest.NewRecorder()

					req, err := http.NewRequest("GET", "/test", nil)
					assert.NoError(t, err)

					req.RemoteAddr = testClientIP + ":1234"
					router.ServeHTTP(w, req)
					assert.Equal(t, tt.expectedCode, w.Code)

					return
				}

				mockRedis := new(MockRedis)
				now := time.Now()

				mockCmd := createMockStringCmd(fmt.Sprintf("%d:%d", tt.tokens, now.Unix()), nil)
				mockRedis.On("Get", mock.Anything, mock.MatchedBy(func(k string) bool {
					return strings.HasPrefix(k, "rate_limit:")
				})).Return(mockCmd, nil)

				mockRedis.On("Set", mock.Anything, mock.MatchedBy(func(k string) bool {
					return strings.HasPrefix(k, "rate_limit:")
				}), mock.Anything, mock.Anything).Return(createMockStatusCmd(nil), nil)

				mockRedis.On("SetWithTTL", mock.Anything, mock.MatchedBy(func(k string) bool {
					return strings.HasPrefix(k, "rate_limit:")
				}), mock.Anything).Return(createMockStatusCmd(nil), nil)

				limiter := NewRateLimiterWithRedis(mockRedis, 5, time.Second)
				limiter.timeNow = func() time.Time { return now }

				router := setupTestRouter(limiter.Middleware())
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/test", nil)
				assert.NoError(t, err)
				req.RemoteAddr = testClientIP + ":1234"

				router.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)
			} else {
				router := setupTestRouter(RateLimit(5, time.Second))
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/test", nil)
				assert.NoError(t, err)
				req.RemoteAddr = testClientIP + ":1234"

				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestRecentOffendersCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRedis := new(MockRedis)
	now := time.Now()
	clientID := "offender-ip"
	key := fmt.Sprintf("rate_limit:%s", clientID)
	mockCmd := createMockStringCmd("0:"+fmt.Sprint(now.Unix()), nil)
	mockRedis.On("Get", mock.Anything, key).Return(mockCmd, nil)
	mockRedis.On("SetWithTTL", mock.Anything, key, mock.Anything).Return(createMockStatusCmd(nil), nil)

	limiter := NewRateLimiterWithRedis(mockRedis, 5, time.Second)
	limiter.timeNow = func() time.Time { return now }
	handler := limiter.Middleware()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = clientID + ":1234"

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = clientID + ":1234"

	start := time.Now()
	router.ServeHTTP(w2, req2)
	duration := time.Since(start)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Less(t, duration.Milliseconds(), int64(10), "Should return almost instantly due to in-memory cache")
}
