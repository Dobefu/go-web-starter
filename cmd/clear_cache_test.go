package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRedisClient struct {
	mock.Mock
	redisClient.Cmdable
}

func (m *mockRedisClient) FlushDB(ctx context.Context) *redisClient.StatusCmd {
	args := m.Called(ctx)
	cmd := redisClient.NewStatusCmd(ctx)

	if err, ok := args.Get(0).(error); ok && err != nil {
		cmd.SetErr(err)
	}

	if val, ok := args.Get(1).(string); ok && val != "" {
		cmd.SetVal(val)
	}

	return cmd
}

func (m *mockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockRedisClient) Ping(ctx context.Context) *redisClient.StatusCmd {
	args := m.Called(ctx)
	cmd := redisClient.NewStatusCmd(ctx)
	cmd.SetErr(args.Error(0))

	return cmd
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestRunClearCacheCmd(t *testing.T) {
	defaultViper := func() {
		viper.Set("redis.enable", true)
		viper.Set("redis.host", "localhost")
		viper.Set("redis.port", 6379)
		viper.Set("redis.password", "")
		viper.Set("redis.db", 0)
	}

	type testCase struct {
		name       string
		viperSetup func()
		mockSetup  func(*mockRedisClient)
		newErr     error
		want       string
	}

	cases := []testCase{
		{
			name:       "disabled",
			viperSetup: func() { viper.Set("redis.enable", false) },
			want:       "Redis is not enabled in configuration",
		},
		{
			name:       "init error",
			viperSetup: defaultViper,
			newErr:     errors.New("init error"),
			want:       "Failed to initialize Redis",
		},
		{
			name:       "flush error",
			viperSetup: defaultViper,
			mockSetup: func(m *mockRedisClient) {
				m.On("Close").Return(nil)
				m.On("FlushDB", mock.Anything).Return(errors.New("flush error"), "")
				m.On("Ping", mock.Anything).Return(nil)
			},
			want: "Failed to clear Redis cache",
		},
		{
			name:       "success",
			viperSetup: defaultViper,
			mockSetup: func(m *mockRedisClient) {
				m.On("Close").Return(nil)
				m.On("FlushDB", mock.Anything).Return(nil, "OK")
				m.On("Ping", mock.Anything).Return(nil)
			},
			want: "Redis cache cleared successfully",
		},
	}

	origNew := redis.New
	t.Cleanup(func() { redis.New = origNew })

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()

			if tc.viperSetup != nil {
				tc.viperSetup()
			}

			if tc.newErr != nil {
				redis.New = func(cfg config.Redis, log *logger.Logger) (*redis.Redis, error) {
					return nil, tc.newErr
				}
			} else if tc.mockSetup != nil {
				mockClient := new(mockRedisClient)
				tc.mockSetup(mockClient)

				redis.New = func(cfg config.Redis, log *logger.Logger) (*redis.Redis, error) {
					return redis.NewWithMockDB(mockClient, log), nil
				}
			}
			output := captureOutput(func() {
				runClearCacheCmd(&cobra.Command{}, []string{})
			})

			assert.Contains(t, output, tc.want)
		})
	}
}
