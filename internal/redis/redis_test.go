package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisTestSuite struct {
	suite.Suite
	ctx    context.Context
	logger *logger.Logger
	client *Redis
}

func (s *RedisTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = logger.New(config.GetLogLevel(), os.Stdout)

	s.client = &Redis{
		db:     redisClient.NewClient(&redisClient.Options{Addr: "localhost:9736"}),
		logger: s.logger,
	}
}

func (s *RedisTestSuite) TearDownTest() {
	if s.client != nil && s.client.db != nil {
		s.client.Close()
	}
}

func (s *RedisTestSuite) TestNew() {
	cfg := config.Redis{
		Host:     "localhost",
		Port:     9736,
		Password: "root",
		DB:       0,
	}

	tests := []struct {
		name    string
		logger  *logger.Logger
		wantErr bool
	}{
		{
			name:    "with logger",
			logger:  s.logger,
			wantErr: false,
		},
		{
			name:    "without logger",
			logger:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		s.Run(tt.name, func() {
			redis, err := New(cfg, tt.logger)

			if tt.wantErr {
				assert.Error(s.T(), err)
				assert.Nil(s.T(), redis)
			} else {
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), redis)
				assert.NotNil(s.T(), redis.db)
				assert.Equal(s.T(), tt.logger, redis.logger)
				redis.Close()
			}
		})
	}
}

func (s *RedisTestSuite) TestRedisOperations() {
	operations := []struct {
		name       string
		op         func(*Redis) error
		wantErr    bool
		skipUninit bool
	}{
		{
			name: "Get non-existing key",
			op: func(r *Redis) error {
				if r.db == nil {
					return errNotInitialized
				}

				_, err := r.Get(s.ctx, "non-existing-key")

				if err != redisClient.Nil {
					return fmt.Errorf("expected redis.Nil error, got %v", err)
				}

				r.logger = nil
				_, err = r.Get(s.ctx, "non-existing-key")

				if err != redisClient.Nil {
					return fmt.Errorf("expected redis.Nil error with nil logger, got %v", err)
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "Basic Redis operations",
			op: func(r *Redis) error {
				if r.db == nil {
					return errNotInitialized
				}

				_, err := r.Set(s.ctx, "test-key", "test-value", time.Hour)

				if err != nil {
					return fmt.Errorf("Set failed: %v", err)
				}

				val, err := r.Get(s.ctx, "test-key")

				if err != nil {
					return fmt.Errorf("Get failed: %v", err)
				}

				if val.Val() != "test-value" {
					return fmt.Errorf("expected 'test-value', got '%s'", val.Val())
				}

				_, err = r.SetRange(s.ctx, "test-key", 5, "-modified")

				if err != nil {
					return fmt.Errorf("SetRange failed: %v", err)
				}

				rangeVal, err := r.GetRange(s.ctx, "test-key", 5, 13)

				if err != nil {
					return fmt.Errorf("GetRange failed: %v", err)
				}

				if rangeVal.Val() != "-modified" {
					return fmt.Errorf("expected '-modified', got '%s'", rangeVal.Val())
				}

				_, err = r.FlushDB(s.ctx)

				if err != nil {
					return fmt.Errorf("FlushDB failed: %v", err)
				}

				r.logger = nil

				_, err = r.Set(s.ctx, "test-key", "test-value", time.Hour)

				if err != nil {
					return fmt.Errorf("Set failed with nil logger: %v", err)
				}

				val, err = r.Get(s.ctx, "test-key")

				if err != nil {
					return fmt.Errorf("Get failed with nil logger: %v", err)
				}

				if val.Val() != "test-value" {
					return fmt.Errorf("expected 'test-value', got '%s'", val.Val())
				}

				_, err = r.SetRange(s.ctx, "test-key", 5, "-modified")

				if err != nil {
					return fmt.Errorf("SetRange failed with nil logger: %v", err)
				}

				rangeVal, err = r.GetRange(s.ctx, "test-key", 5, 13)

				if err != nil {
					return fmt.Errorf("GetRange failed with nil logger: %v", err)
				}

				if rangeVal.Val() != "-modified" {
					return fmt.Errorf("expected '-modified', got '%s'", rangeVal.Val())
				}

				_, err = r.FlushDB(s.ctx)

				if err != nil {
					return fmt.Errorf("FlushDB failed with nil logger: %v", err)
				}

				_, err = r.Get(s.ctx, "test-key")

				if err != redisClient.Nil {
					return fmt.Errorf("expected key to be removed after FlushDB, got %v", err)
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "Error handling with invalid connection",
			op: func(r *Redis) error {
				errorClient := &Redis{
					db:     redisClient.NewClient(&redisClient.Options{Addr: "localhost:1"}),
					logger: s.logger,
				}

				defer errorClient.Close()

				_, err := errorClient.Set(s.ctx, "test-key", "test-value", time.Hour)

				if err == nil {
					return fmt.Errorf("expected error for Set with invalid client")
				}

				_, err = errorClient.Get(s.ctx, "test-key")

				if err == nil {
					return fmt.Errorf("expected error for Get with invalid client")
				}

				_, err = errorClient.GetRange(s.ctx, "test-key", 0, 5)

				if err == nil {
					return fmt.Errorf("expected error for GetRange with invalid client")
				}

				_, err = errorClient.SetRange(s.ctx, "test-key", 0, "modified")

				if err == nil {
					return fmt.Errorf("expected error for SetRange with invalid client")
				}

				_, err = errorClient.FlushDB(s.ctx)

				if err == nil {
					return fmt.Errorf("expected error for FlushDB with invalid client")
				}

				errorClient.logger = nil

				_, err = errorClient.Set(s.ctx, "test-key", "test-value", time.Hour)

				if err == nil {
					return fmt.Errorf("expected error for Set with invalid client and nil logger")
				}

				_, err = errorClient.Get(s.ctx, "test-key")

				if err == nil {
					return fmt.Errorf("expected error for Get with invalid client and nil logger")
				}

				_, err = errorClient.GetRange(s.ctx, "test-key", 0, 5)

				if err == nil {
					return fmt.Errorf("expected error for GetRange with invalid client and nil logger")
				}

				_, err = errorClient.SetRange(s.ctx, "test-key", 0, "modified")

				if err == nil {
					return fmt.Errorf("expected error for SetRange with invalid client and nil logger")
				}

				_, err = errorClient.FlushDB(s.ctx)

				if err == nil {
					return fmt.Errorf("expected error for FlushDB with invalid client and nil logger")
				}

				return nil
			},
			wantErr:    false,
			skipUninit: true,
		},
		{
			name: "Closed client operations",
			op: func(r *Redis) error {
				closedClient := &Redis{
					db:     redisClient.NewClient(&redisClient.Options{Addr: "localhost:9736"}),
					logger: s.logger,
				}

				closedClient.Close()

				_, err := closedClient.Set(s.ctx, "test-key", "test-value", time.Hour)

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for Set, got %v", err)
				}

				_, err = closedClient.Get(s.ctx, "test-key")

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for Get, got %v", err)
				}

				_, err = closedClient.GetRange(s.ctx, "test-key", 0, 5)

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for GetRange, got %v", err)
				}

				_, err = closedClient.SetRange(s.ctx, "test-key", 0, "modified")

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for SetRange, got %v", err)
				}

				_, err = closedClient.FlushDB(s.ctx)

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for FlushDB, got %v", err)
				}

				closedClient.logger = nil

				_, err = closedClient.Set(s.ctx, "test-key", "test-value", time.Hour)

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for Set with nil logger, got %v", err)
				}

				_, err = closedClient.Get(s.ctx, "test-key")

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for Get with nil logger, got %v", err)
				}

				_, err = closedClient.GetRange(s.ctx, "test-key", 0, 5)

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for GetRange with nil logger, got %v", err)
				}

				_, err = closedClient.SetRange(s.ctx, "test-key", 0, "modified")

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for SetRange with nil logger, got %v", err)
				}

				_, err = closedClient.FlushDB(s.ctx)

				if err != errClientClosed {
					return fmt.Errorf("expected errClientClosed for FlushDB with nil logger, got %v", err)
				}

				return nil
			},
			wantErr:    false,
			skipUninit: true,
		},
		{
			name: "Operations with uninitialized client",
			op: func(r *Redis) error {
				uninitClient := &Redis{
					db:     nil,
					logger: s.logger,
				}

				if _, err := uninitClient.Get(s.ctx, "key"); err != errNotInitialized {
					return fmt.Errorf("expected errNotInitialized for Get, got %v", err)
				}

				if _, err := uninitClient.Set(s.ctx, "key", "value", time.Hour); err != errNotInitialized {
					return fmt.Errorf("expected errNotInitialized for Set, got %v", err)
				}

				if _, err := uninitClient.GetRange(s.ctx, "key", 0, 5); err != errNotInitialized {
					return fmt.Errorf("expected errNotInitialized for GetRange, got %v", err)
				}

				if _, err := uninitClient.SetRange(s.ctx, "key", 0, "value"); err != errNotInitialized {
					return fmt.Errorf("expected errNotInitialized for SetRange, got %v", err)
				}

				if _, err := uninitClient.FlushDB(s.ctx); err != errNotInitialized {
					return fmt.Errorf("expected errNotInitialized for FlushDB, got %v", err)
				}

				return nil
			},
			wantErr:    false,
			skipUninit: true,
		},
	}

	for _, op := range operations {
		op := op

		s.Run(op.name, func() {
			redis := s.client
			err := op.op(redis)

			if op.wantErr {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})

		if !op.skipUninit {
			s.Run(op.name+"_uninitialized", func() {
				uninitRedis := &Redis{db: nil, logger: s.logger}
				err := op.op(uninitRedis)
				assert.Equal(s.T(), errNotInitialized, err)

				uninitRedis.logger = nil
				err = op.op(uninitRedis)
				assert.Equal(s.T(), errNotInitialized, err)
			})
		}
	}
}

func (s *RedisTestSuite) TestIsClientClosed() {
	tests := []struct {
		name     string
		setup    func() *Redis
		expected bool
	}{
		{
			name: "nil client",
			setup: func() *Redis {
				return &Redis{db: nil}
			},
			expected: false,
		},
		{
			name: "closed client",
			setup: func() *Redis {
				client := &Redis{
					db: redisClient.NewClient(&redisClient.Options{Addr: "localhost:9736"}),
				}
				client.Close()
				return client
			},
			expected: true,
		},
		{
			name: "active client",
			setup: func() *Redis {
				return &Redis{
					db: redisClient.NewClient(&redisClient.Options{Addr: "localhost:9736"}),
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			client := tt.setup()
			result := client.isClientClosed()
			assert.Equal(s.T(), tt.expected, result)

			if client.db != nil {
				client.Close()
			}
		})
	}
}

func (s *RedisTestSuite) TestClose() {
	tests := []struct {
		name    string
		client  *Redis
		wantErr error
	}{
		{
			name:    "nil client",
			client:  &Redis{db: nil},
			wantErr: errNotInitialized,
		},
		{
			name: "valid client",
			client: &Redis{
				db: redisClient.NewClient(&redisClient.Options{Addr: "localhost:9736"}),
			},
			wantErr: nil,
		},
		{
			name: "already closed client",
			client: func() *Redis {
				client := &Redis{
					db: redisClient.NewClient(&redisClient.Options{Addr: "localhost:9736"}),
				}
				client.Close()
				return client
			}(),
			wantErr: errClientClosed,
		},
	}

	for _, tt := range tests {
		tt := tt

		s.Run(tt.name, func() {
			err := tt.client.Close()
			if tt.wantErr != nil {
				assert.Equal(s.T(), tt.wantErr, err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
