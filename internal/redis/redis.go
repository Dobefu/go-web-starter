package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	redis "github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Close() error
	Get(ctx context.Context, key string) (*redis.StringCmd, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) (*redis.StatusCmd, error)
	GetRange(ctx context.Context, key string, start, end int64) (*redis.StringCmd, error)
	SetRange(ctx context.Context, key string, offset int64, value string) (*redis.IntCmd, error)
}

type Redis struct {
	db     *redis.Client
	logger *logger.Logger
}

var errNotInitialized error = fmt.Errorf("redis not initialized")

var New = func(cfg config.Redis, log *logger.Logger) (*Redis, error) {
	db := redis.NewClient(&redis.Options{
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

func (d *Redis) Get(ctx context.Context, key string) (*redis.StringCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.Get(ctx, key), nil
}

func (d *Redis) Set(ctx context.Context, key string, value any, expiration time.Duration) (*redis.StatusCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.Set(ctx, key, value, expiration), nil
}

func (d *Redis) GetRange(ctx context.Context, key string, start, end int64) (*redis.StringCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.GetRange(ctx, key, start, end), nil
}

func (d *Redis) SetRange(ctx context.Context, key string, offset int64, value string) (*redis.IntCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	return d.db.SetRange(ctx, key, offset, value), nil
}
