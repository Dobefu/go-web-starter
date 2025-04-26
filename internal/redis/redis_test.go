package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRedisClient struct {
	mock.Mock
	redisClient.Cmdable
}

func (m *mockRedisClient) Ping(ctx context.Context) *redisClient.StatusCmd {
	args := m.Called(ctx)
	cmd := redisClient.NewStatusCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redisClient.StringCmd {
	args := m.Called(ctx, key)
	cmd := redisClient.NewStringCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redisClient.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	cmd := redisClient.NewStatusCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func (m *mockRedisClient) GetRange(ctx context.Context, key string, start, end int64) *redisClient.StringCmd {
	args := m.Called(ctx, key, start, end)
	cmd := redisClient.NewStringCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func (m *mockRedisClient) SetRange(ctx context.Context, key string, offset int64, value string) *redisClient.IntCmd {
	args := m.Called(ctx, key, offset, value)
	cmd := redisClient.NewIntCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func (m *mockRedisClient) FlushDB(ctx context.Context) *redisClient.StatusCmd {
	args := m.Called(ctx)
	cmd := redisClient.NewStatusCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func (m *mockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type pingErrClient struct {
	mockRedisClient
}

func (p *pingErrClient) Ping(ctx context.Context) *redisClient.StatusCmd {
	cmd := redisClient.NewStatusCmd(ctx)
	cmd.SetErr(errors.New("some other error"))

	return cmd
}

func TestRedis_Close(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(*mockRedisClient)
		nilDB     bool
		expectErr error
	}

	cases := []testCase{
		{
			name:      "nil db",
			nilDB:     true,
			expectErr: errNotInitialized,
		},
		{
			name: "close ok",
			setupMock: func(m *mockRedisClient) {
				m.On("Close").Return(nil)
			},
			expectErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r *Redis

			if tc.nilDB {
				r = &Redis{db: nil, logger: newTestLogger(t)}
			} else {
				mockClient := new(mockRedisClient)

				if tc.setupMock != nil {
					tc.setupMock(mockClient)
				}

				r = newTestRedis(mockClient, t)
			}
			err := r.Close()

			if tc.expectErr != nil {
				assert.Equal(t, tc.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newTestLogger(t *testing.T) *logger.Logger {
	return logger.New(logger.DebugLevel, testingWriter{t})
}

func newTestRedis(mockClient *mockRedisClient, t *testing.T) *Redis {
	return &Redis{db: mockClient, logger: newTestLogger(t)}
}

func newTestContext() context.Context {
	return context.Background()
}

type redisTestCase struct {
	name      string
	setupMock func(*mockRedisClient)
	nilDB     bool
	call      func(r *Redis) (any, error)
	expectNil bool
	expectErr error
}

func runRedisMethodTests(t *testing.T, cases []redisTestCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r *Redis

			if tc.nilDB {
				r = &Redis{db: nil, logger: newTestLogger(t)}
			} else {
				mockClient := new(mockRedisClient)

				if tc.setupMock != nil {
					tc.setupMock(mockClient)
				}

				r = newTestRedis(mockClient, t)
			}
			result, err := tc.call(r)

			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}

			if tc.expectErr != nil {
				assert.Equal(t, tc.expectErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedis_Get(t *testing.T) {
	key := "test-key"
	runRedisMethodTests(t, []redisTestCase{
		{
			name:      "nil db",
			nilDB:     true,
			call:      func(r *Redis) (any, error) { cmd, err := r.Get(newTestContext(), key); return cmd, err },
			expectNil: true,
			expectErr: errNotInitialized,
		},
		{
			name: "closed client",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(errors.New("redis: client is closed"))
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.Get(newTestContext(), key); return cmd, err },
			expectNil: true,
			expectErr: errClientClosed,
		},
		{
			name: "key not found",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(nil)
				m.On("Get", mock.Anything, key).Return(redisClient.Nil)
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.Get(newTestContext(), key); return cmd, err },
			expectNil: true,
			expectErr: redisClient.Nil,
		},
		{
			name: "other error",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(nil)
				m.On("Get", mock.Anything, key).Return(errors.New("some error"))
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.Get(newTestContext(), key); return cmd, err },
			expectNil: true,
			expectErr: errors.New("some error"),
		},
	})
}

func TestRedis_Set(t *testing.T) {
	key := "test-key"
	val := "val"

	runRedisMethodTests(t, []redisTestCase{
		{
			name:      "nil db",
			nilDB:     true,
			call:      func(r *Redis) (any, error) { cmd, err := r.Set(newTestContext(), key, val, 0); return cmd, err },
			expectNil: true,
			expectErr: errNotInitialized,
		},
		{
			name: "closed client",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(errors.New("redis: client is closed"))
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.Set(newTestContext(), key, val, 0); return cmd, err },
			expectNil: true,
			expectErr: errClientClosed,
		},
		{
			name: "set error",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(nil)
				m.On("Set", mock.Anything, key, val, time.Duration(0)).Return(errors.New("set error"))
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.Set(newTestContext(), key, val, 0); return cmd, err },
			expectNil: false,
			expectErr: errors.New("set error"),
		},
	})
}

func TestRedis_GetRange(t *testing.T) {
	key := "test-key"
	start, end := int64(0), int64(10)

	runRedisMethodTests(t, []redisTestCase{
		{
			name:  "nil db",
			nilDB: true,
			call: func(r *Redis) (any, error) {
				cmd, err := r.GetRange(newTestContext(), key, start, end)
				return cmd, err
			},
			expectNil: true,
			expectErr: errNotInitialized,
		},
		{
			name: "closed client",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(errors.New("redis: client is closed"))
			},
			call: func(r *Redis) (any, error) {
				cmd, err := r.GetRange(newTestContext(), key, start, end)
				return cmd, err
			},
			expectNil: true,
			expectErr: errClientClosed,
		},
		{
			name: "getrange error",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(nil)
				m.On("GetRange", mock.Anything, key, start, end).Return(errors.New("getrange error"))
			},
			call: func(r *Redis) (any, error) {
				cmd, err := r.GetRange(newTestContext(), key, start, end)
				return cmd, err
			},
			expectNil: false,
			expectErr: errors.New("getrange error"),
		},
	})
}

func TestRedis_SetRange(t *testing.T) {
	key := "test-key"

	runRedisMethodTests(t, []redisTestCase{
		{
			name:  "nil db",
			nilDB: true,
			call: func(r *Redis) (any, error) {
				cmd, err := r.SetRange(newTestContext(), key, 0, "val")
				return cmd, err
			},
			expectNil: true,
			expectErr: errNotInitialized,
		},
		{
			name: "closed client",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(errors.New("redis: client is closed"))
			},
			call: func(r *Redis) (any, error) {
				cmd, err := r.SetRange(newTestContext(), key, 0, "val")
				return cmd, err
			},
			expectNil: true,
			expectErr: errClientClosed,
		},
		{
			name: "setrange error",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(nil)
				m.On("SetRange", mock.Anything, key, int64(0), "val").Return(errors.New("setrange error"))
			},
			call: func(r *Redis) (any, error) {
				cmd, err := r.SetRange(newTestContext(), key, 0, "val")
				return cmd, err
			},
			expectNil: false,
			expectErr: errors.New("setrange error"),
		},
	})
}

func TestRedis_FlushDB(t *testing.T) {
	runRedisMethodTests(t, []redisTestCase{
		{
			name:      "nil db",
			nilDB:     true,
			call:      func(r *Redis) (any, error) { cmd, err := r.FlushDB(newTestContext()); return cmd, err },
			expectNil: true,
			expectErr: errNotInitialized,
		},
		{
			name: "closed client",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(errors.New("redis: client is closed"))
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.FlushDB(newTestContext()); return cmd, err },
			expectNil: true,
			expectErr: errClientClosed,
		},
		{
			name: "flushdb error",
			setupMock: func(m *mockRedisClient) {
				m.On("Ping", mock.Anything).Return(nil)
				m.On("FlushDB", mock.Anything).Return(errors.New("flushdb error"))
			},
			call:      func(r *Redis) (any, error) { cmd, err := r.FlushDB(newTestContext()); return cmd, err },
			expectNil: false,
			expectErr: errors.New("flushdb error"),
		},
	})
}

func TestRedis_isClientClosed(t *testing.T) {
	type testCase struct {
		name      string
		redis     *Redis
		expectVal bool
	}

	log := newTestLogger(t)
	cases := []testCase{
		{
			name:      "db is nil",
			redis:     &Redis{db: nil, logger: log},
			expectVal: false,
		},
		{
			name:      "Ping returns error not containing 'client is closed'",
			redis:     &Redis{db: &pingErrClient{}, logger: log},
			expectVal: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectVal, tc.redis.isClientClosed())
		})
	}
}

func TestRedis_New(t *testing.T) {
	log := logger.New(logger.DebugLevel, testingWriter{t})

	cfg := config.Redis{
		Enable:   true,
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	newFunc := New

	if newFunc == nil {
		t.Fatal("New function is nil")
	}

	_, err := New(cfg, log)
	assert.NoError(t, err)
}

type testingWriter struct{ t *testing.T }

func (tw testingWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}
