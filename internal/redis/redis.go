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
	FlushDB(ctx context.Context) (*redisClient.StatusCmd, error)
}

type Redis struct {
	db     *redisClient.Client
	logger *logger.Logger
}

var errNotInitialized error = fmt.Errorf("redis not initialized")

var New = func(cfg config.Redis, log *logger.Logger) (*Redis, error) {
	db := redisClient.NewClient(&redisClient.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
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

	if d.logger != nil {
		d.logger.Debug("Executing Redis GET", logger.Fields{
			"key": key,
		})
	}

	cmd := d.db.Get(ctx, key)

	if cmd.Err() != nil && d.logger != nil {
		d.logger.Error("Redis GET failed", logger.Fields{
			"key":   key,
			"error": cmd.Err().Error(),
		})
	}

	return cmd, nil
}

func (d *Redis) Set(ctx context.Context, key string, value any, expiration time.Duration) (*redisClient.StatusCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	if d.logger != nil {
		d.logger.Debug("Executing Redis SET", logger.Fields{
			"key":        key,
			"expiration": expiration,
		})
	}

	cmd := d.db.Set(ctx, key, value, expiration)

	if cmd.Err() != nil && d.logger != nil {
		d.logger.Error("Redis SET failed", logger.Fields{
			"key":   key,
			"error": cmd.Err().Error(),
		})
	}

	return cmd, nil
}

func (d *Redis) GetRange(ctx context.Context, key string, start, end int64) (*redisClient.StringCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	if d.logger != nil {
		d.logger.Debug("Executing Redis GETRANGE", logger.Fields{
			"key":   key,
			"start": start,
			"end":   end,
		})
	}

	cmd := d.db.GetRange(ctx, key, start, end)

	if cmd.Err() != nil && d.logger != nil {
		d.logger.Error("Redis GETRANGE failed", logger.Fields{
			"key":   key,
			"error": cmd.Err().Error(),
		})
	}

	return cmd, nil
}

func (d *Redis) SetRange(ctx context.Context, key string, offset int64, value string) (*redisClient.IntCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	if d.logger != nil {
		d.logger.Debug("Executing Redis SETRANGE", logger.Fields{
			"key":    key,
			"offset": offset,
		})
	}

	cmd := d.db.SetRange(ctx, key, offset, value)

	if cmd.Err() != nil && d.logger != nil {
		d.logger.Error("Redis SETRANGE failed", logger.Fields{
			"key":   key,
			"error": cmd.Err().Error(),
		})
	}

	return cmd, nil
}

func (d *Redis) FlushDB(ctx context.Context) (*redisClient.StatusCmd, error) {
	if d.db == nil {
		return nil, errNotInitialized
	}

	if d.logger != nil {
		d.logger.Info("Executing Redis FLUSHDB", nil)
	}

	cmd := d.db.FlushDB(ctx)

	if cmd.Err() != nil && d.logger != nil {
		d.logger.Error("Redis FLUSHDB failed", logger.Fields{
			"error": cmd.Err().Error(),
		})
	}

	return cmd, nil
}
