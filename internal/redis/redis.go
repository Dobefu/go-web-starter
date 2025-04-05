package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	redisClient "github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Close() error
	Get(ctx context.Context, key string) (*redisClient.StringCmd, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) (*redisClient.StatusCmd, error)
	GetRange(ctx context.Context, key string, start, end int64) (*redisClient.StringCmd, error)
	SetRange(ctx context.Context, key string, offset int64, value string) (*redisClient.IntCmd, error)
}

type Redis struct {
	db     *redisClient.Client
	logger *logger.Logger
}

var errNotInitialized error = fmt.Errorf("redis not initialized")

var New = func(cfg config.Redis, log *logger.Logger) (*Redis, error) {
	db := redisClient.NewClient(&redisClient.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Redis{
		db:     db,
		logger: log,
	}, nil
}

func (d *Redis) Close() error {
	if d.db == nil {
		return errNotInitialized
	}

	return d.db.Close()
}

func (d *Redis) Get(ctx context.Context, key string) (*redisClient.StringCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.Get(ctx, key), nil
}

func (d *Redis) Set(ctx context.Context, key string, value any, expiration time.Duration) (*redisClient.StatusCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.Set(ctx, key, value, expiration), nil
}

func (d *Redis) GetRange(ctx context.Context, key string, start, end int64) (*redisClient.StringCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.GetRange(ctx, key, start, end), nil
}

func (d *Redis) SetRange(ctx context.Context, key string, offset int64, value string) (*redisClient.IntCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.SetRange(ctx, key, offset, value), nil
}
