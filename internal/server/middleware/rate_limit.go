package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	redis    redis.RedisInterface
	capacity int
	rate     time.Duration
	logger   *logger.Logger
	timeNow  func() time.Time
}

func NewRateLimiter(capacity int, rate time.Duration) (*RateLimiter, error) {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	log.Debug("Initializing rate limiter", logger.Fields{
		"capacity": capacity,
		"rate":     rate.String(),
	})

	cfg := config.DefaultConfig.Redis
	redisClient, err := redis.New(cfg, log)

	if err != nil {
		log.Error("Failed to create Redis client for rate limiter", logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	return &RateLimiter{
		redis:    redisClient,
		capacity: capacity,
		rate:     rate,
		logger:   log,
		timeNow:  time.Now,
	}, nil
}

func NewRateLimiterWithRedis(redisClient redis.RedisInterface, capacity int, rate time.Duration) *RateLimiter {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	log.Info("Initializing rate limiter with existing Redis client", logger.Fields{
		"capacity": capacity,
		"rate":     rate.String(),
	})

	return &RateLimiter{
		redis:    redisClient,
		capacity: capacity,
		rate:     rate,
		logger:   log,
		timeNow:  time.Now,
	}
}

func getClientIP(c *gin.Context) string {
	if forwardedIp := c.GetHeader("X-Forwarded-For"); forwardedIp != "" {
		ips := strings.Split(forwardedIp, ",")

		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
		return xrip
	}

	if c.Request.RemoteAddr != "" {
		return strings.Split(c.Request.RemoteAddr, ":")[0]
	}

	return "unknown"
}

func (rl *RateLimiter) Allow(clientID string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", clientID)

	val, err := rl.redis.Get(ctx, key)
	now := rl.timeNow()

	var tokens int
	var lastUpdate time.Time

	if err != nil && err.Error() == "redis: nil" {
		tokens = rl.capacity
		lastUpdate = now

		rl.logger.Debug("New rate limit entry", logger.Fields{
			"tokens":     tokens,
			"lastUpdate": lastUpdate,
			"key":        key,
		})
	} else if err != nil {
		rl.logger.Error("Failed to get rate limit data", logger.Fields{
			"error": err.Error(),
			"key":   key,
		})

		return true
	} else {
		var lastUpdateUnix int64
		_, err = fmt.Sscanf(val.Val(), "%d:%d", &tokens, &lastUpdateUnix)

		if err != nil {
			rl.logger.Error("Failed to parse rate limit data", logger.Fields{
				"error": err.Error(),
				"value": val.Val(),
			})

			return true
		}

		lastUpdate = time.Unix(lastUpdateUnix, 0)
		elapsed := now.Sub(lastUpdate)
		tokensToAdd := int(elapsed / rl.rate)

		if tokensToAdd > 0 {
			tokens = int(math.Min(float64(rl.capacity), float64(tokens+tokensToAdd)))
			lastUpdate = now

			rl.logger.Debug("Refilled tokens", logger.Fields{
				"tokensToAdd": tokensToAdd,
				"newTokens":   tokens,
				"key":         key,
			})
		}
	}

	if tokens <= 0 {
		value := fmt.Sprintf("%d:%d", tokens, lastUpdate.Unix())
		_, err = rl.redis.Set(ctx, key, value, rl.rate*2)

		if err != nil {
			rl.logger.Error("Failed to update rate limit data", logger.Fields{
				"error": err.Error(),
				"key":   key,
			})
		}

		rl.logger.Debug("Rate limit exceeded", logger.Fields{
			"tokens": tokens,
			"key":    key,
		})

		return false
	}

	tokens -= 1
	value := fmt.Sprintf("%d:%d", tokens, lastUpdate.Unix())
	_, err = rl.redis.Set(ctx, key, value, rl.rate*2)

	if err != nil {
		rl.logger.Error("Failed to update rate limit data", logger.Fields{
			"error": err.Error(),
			"key":   key,
		})

		return true
	}

	rl.logger.Debug("Rate limit updated", logger.Fields{
		"remainingTokens": tokens,
		"key":             key,
	})

	return true
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow(getClientIP(c)) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})

			c.Abort()

			return
		}

		c.Next()
	}
}

func RateLimit(capacity int, rate time.Duration) gin.HandlerFunc {
	if !config.DefaultConfig.Redis.Enable {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter, err := NewRateLimiter(capacity, rate)

	if err != nil {
		panic(fmt.Sprintf("Failed to create rate limiter: %v", err))
	}

	return func(c *gin.Context) {
		clientIP := getClientIP(c)

		if !limiter.Allow(clientIP) {
			limiter.logger.Warn("Rate limit exceeded", logger.Fields{
				"client_ip": clientIP,
				"path":      c.Request.URL.Path,
			})

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})

			c.Abort()
			return
		}

		c.Next()
	}
}
