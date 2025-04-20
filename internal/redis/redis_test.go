package redis

import (
	"context"
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
}

func (s *RedisTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = logger.New(config.GetLogLevel(), os.Stdout)
}

func (s *RedisTestSuite) newRedisClient() *Redis {
	return &Redis{
		db:     redisClient.NewClient(&redisClient.Options{Addr: "localhost:6379"}),
		logger: s.logger,
	}
}

func (s *RedisTestSuite) TestNew() {
	cfg := config.Redis{
		Host:     "localhost",
		Port:     6379,
		Password: "password",
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
			s.T().Parallel()

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

func (s *RedisTestSuite) TestClose() {
	tests := []struct {
		name    string
		setup   func() *Redis
		wantErr bool
	}{
		{
			name: "success",
			setup: func() *Redis {
				return s.newRedisClient()
			},
			wantErr: false,
		},
		{
			name: "already closed",
			setup: func() *Redis {
				redis := s.newRedisClient()
				redis.db.Close()

				return redis
			},
			wantErr: true,
		},
		{
			name: "uninitialized",
			setup: func() *Redis {
				return &Redis{db: nil}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		s.Run(tt.name, func() {
			s.T().Parallel()

			redis := tt.setup()
			err := redis.Close()

			if tt.wantErr {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})
	}
}

func (s *RedisTestSuite) TestRedisOperations() {
	operations := []struct {
		name    string
		op      func(*Redis) error
		wantErr bool
	}{
		{
			name: "Get",
			op: func(r *Redis) error {
				_, err := r.Get(s.ctx, "test-key")
				return err
			},
			wantErr: false,
		},
		{
			name: "Set",
			op: func(r *Redis) error {
				_, err := r.Set(s.ctx, "test-key", "test-value", time.Hour)
				return err
			},
			wantErr: false,
		},
		{
			name: "GetRange",
			op: func(r *Redis) error {
				_, err := r.GetRange(s.ctx, "test-key", 0, 10)
				return err
			},
			wantErr: false,
		},
		{
			name: "SetRange",
			op: func(r *Redis) error {
				_, err := r.SetRange(s.ctx, "test-key", 0, "test-value")
				return err
			},
			wantErr: false,
		},
		{
			name: "FlushDB",
			op: func(r *Redis) error {
				_, err := r.FlushDB(s.ctx)
				return err
			},
			wantErr: false,
		},
	}

	for _, op := range operations {
		op := op

		s.Run(op.name, func() {
			s.T().Parallel()

			redis := s.newRedisClient()
			defer redis.Close()
			err := op.op(redis)

			if op.wantErr {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})

		s.Run(op.name+" uninitialized", func() {
			s.T().Parallel()

			uninitRedis := &Redis{db: nil, logger: s.logger}
			err := op.op(uninitRedis)
			assert.Error(s.T(), err)
			assert.Equal(s.T(), errNotInitialized, err)
		})
	}
}

func (s *RedisTestSuite) TestErrorLogging() {
	operations := []struct {
		name string
		op   func(*Redis) error
	}{
		{
			name: "Get error",
			op: func(r *Redis) error {
				_, err := r.Get(s.ctx, "error-key")
				return err
			},
		},
		{
			name: "Set error",
			op: func(r *Redis) error {
				_, err := r.Set(s.ctx, "error-key", "error-value", time.Hour)
				return err
			},
		},
		{
			name: "GetRange error",
			op: func(r *Redis) error {
				_, err := r.GetRange(s.ctx, "error-key", -1, -1)
				return err
			},
		},
		{
			name: "SetRange error",
			op: func(r *Redis) error {
				_, err := r.SetRange(s.ctx, "error-key", -1, "error-value")
				return err
			},
		},
		{
			name: "FlushDB error",
			op: func(r *Redis) error {
				_, err := r.FlushDB(s.ctx)
				return err
			},
		},
	}

	for _, op := range operations {
		op := op

		s.Run(op.name, func() {
			s.T().Parallel()
			redis := s.newRedisClient()
			redis.db.Close()
			err := op.op(redis)
			assert.NoError(s.T(), err)
		})
	}
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
