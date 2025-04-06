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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

	return args.Get(0).(*redisClient.StringCmd), args.Error(1)
}

func (m *MockRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) (*redisClient.StatusCmd, error) {
	args := m.Called(ctx, key, value, expiration)

	return args.Get(0).(*redisClient.StatusCmd), args.Error(1)
}

func (m *MockRedis) GetRange(ctx context.Context, key string, start, end int64) (*redisClient.StringCmd, error) {
	args := m.Called(ctx, key, start, end)

	return args.Get(0).(*redisClient.StringCmd), args.Error(1)
}

func (m *MockRedis) SetRange(ctx context.Context, key string, offset int64, value string) (*redisClient.IntCmd, error) {
	args := m.Called(ctx, key, offset, value)

	return args.Get(0).(*redisClient.IntCmd), args.Error(1)
}

type RedisTestSuite struct {
	suite.Suite
	mockRedis *MockRedis
	realRedis *Redis
	ctx       context.Context
}

func (s *RedisTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mockRedis = new(MockRedis)

	log := logger.New(logger.InfoLevel, os.Stdout)

	s.realRedis = &Redis{
		db:     redisClient.NewClient(&redisClient.Options{}),
		logger: log,
	}
}

func (s *RedisTestSuite) testUninitializedRedis(operation func(*Redis) error) {
	uninitRedis := &Redis{db: nil}
	err := operation(uninitRedis)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), errNotInitialized, err)
}

func (s *RedisTestSuite) testRedisOperation(
	name string,
	mockSetup func(),
	mockOperation func() (any, error),
	realOperation func() (any, error),
	expectNilResult bool,
) {
	s.Run(name, func() {
		mockSetup()

		mockResult, err := mockOperation()
		assert.NoError(s.T(), err)

		if !expectNilResult {
			assert.NotNil(s.T(), mockResult)
		}

		s.mockRedis.AssertExpectations(s.T())

		realResult, err := realOperation()
		assert.NoError(s.T(), err)

		if !expectNilResult {
			assert.NotNil(s.T(), realResult)
		}
	})
}

func (s *RedisTestSuite) TestRedisOperations() {
	s.testRedisOperation(
		"Close",
		func() { s.mockRedis.On("Close").Return(nil) },
		func() (any, error) { return nil, s.mockRedis.Close() },
		func() (any, error) { return nil, s.realRedis.Close() },
		true,
	)

	s.testUninitializedRedis(func(r *Redis) error { return r.Close() })

	key := "test-key"
	mockStringCmd := redisClient.NewStringCmd(s.ctx)

	s.testRedisOperation(
		"Get",
		func() { s.mockRedis.On("Get", s.ctx, key).Return(mockStringCmd, nil) },
		func() (any, error) { return s.mockRedis.Get(s.ctx, key) },
		func() (any, error) { return s.realRedis.Get(s.ctx, key) },
		false,
	)

	s.testUninitializedRedis(func(r *Redis) error { _, err := r.Get(s.ctx, key); return err })

	value := "test-value"
	expiration := time.Hour
	mockStatusCmd := redisClient.NewStatusCmd(s.ctx)

	s.testRedisOperation(
		"Set",
		func() { s.mockRedis.On("Set", s.ctx, key, value, expiration).Return(mockStatusCmd, nil) },
		func() (any, error) { return s.mockRedis.Set(s.ctx, key, value, expiration) },
		func() (any, error) { return s.realRedis.Set(s.ctx, key, value, expiration) },
		false,
	)
	s.testUninitializedRedis(func(r *Redis) error { _, err := r.Set(s.ctx, key, value, expiration); return err })

	start, end := int64(0), int64(10)

	s.testRedisOperation(
		"GetRange",
		func() { s.mockRedis.On("GetRange", s.ctx, key, start, end).Return(mockStringCmd, nil) },
		func() (any, error) { return s.mockRedis.GetRange(s.ctx, key, start, end) },
		func() (any, error) { return s.realRedis.GetRange(s.ctx, key, start, end) },
		false,
	)
	s.testUninitializedRedis(func(r *Redis) error { _, err := r.GetRange(s.ctx, key, start, end); return err })

	offset := int64(0)
	mockIntCmd := redisClient.NewIntCmd(s.ctx)

	s.testRedisOperation(
		"SetRange",
		func() { s.mockRedis.On("SetRange", s.ctx, key, offset, value).Return(mockIntCmd, nil) },
		func() (any, error) { return s.mockRedis.SetRange(s.ctx, key, offset, value) },
		func() (any, error) { return s.realRedis.SetRange(s.ctx, key, offset, value) },
		false,
	)

	s.testUninitializedRedis(func(r *Redis) error { _, err := r.SetRange(s.ctx, key, offset, value); return err })
}

func (s *RedisTestSuite) TestNew() {
	cfg := config.Redis{
		Host:     "localhost",
		Port:     6379,
		Password: "password",
		DB:       0,
	}

	log := logger.New(logger.InfoLevel, os.Stdout)

	redis, err := New(cfg, log)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), redis)
	assert.NotNil(s.T(), redis.db)
	assert.Equal(s.T(), log, redis.logger)
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
