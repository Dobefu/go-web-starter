package middleware

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	redisClient "github.com/redis/go-redis/v9"
)

const (
	errRateLimitExceeded = "Rate limit exceeded"
	errRedisNil          = "redis: nil"
)

type RateLimiter struct {
	redis    redis.RedisInterface
	capacity int
	rate     time.Duration
	logger   *logger.Logger
	timeNow  func() time.Time

	recentOffenders sync.Map
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
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)

		if err == nil {
			return host
		}

		return c.Request.RemoteAddr
	}

	return "unknown"
}

func getClientID(c *gin.Context) string {
	if gin.Mode() == gin.DebugMode {
		return "localdev"
	}

	session := sessions.Default(c)

	if userID := session.Get("userID"); userID != nil {
		return fmt.Sprintf("user:%v", userID)
	}

	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("apiKey:%s", apiKey)
	}

	return getClientIP(c)
}

func (rl *RateLimiter) Allow(clientID string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", clientID)

	val, err := rl.redis.Get(ctx, key)
	now := rl.timeNow()

	tokens, lastUpdate, parseErr := rl.getTokensAndLastUpdate(val, err, now, key)

	if parseErr != nil {
		return true
	}

	if rl.isRateLimited(tokens, lastUpdate, err, ctx, key) {
		return false
	}

	rl.consumeTokenAndPersist(tokens, lastUpdate, err, ctx, key)
	return true
}

func (rl *RateLimiter) getTokensAndLastUpdate(val *redisClient.StringCmd, err error, now time.Time, key string) (int, time.Time, error) {
	var tokens int
	var lastUpdate time.Time

	if err != nil && err.Error() == errRedisNil {
		tokens = rl.capacity
		lastUpdate = now

		rl.logger.Debug("New rate limit entry", logger.Fields{
			"tokens":     tokens,
			"lastUpdate": lastUpdate,
			"key":        key,
		})

		return tokens, lastUpdate, nil
	} else if err != nil {
		rl.logger.Error("Failed to get rate limit data", logger.Fields{
			"error": err.Error(),
			"key":   key,
		})

		return 0, time.Time{}, err
	}

	var lastUpdateUnix int64
	_, scanErr := fmt.Sscanf(val.Val(), "%d:%d", &tokens, &lastUpdateUnix)

	if scanErr != nil {
		rl.logger.Error("Failed to parse rate limit data", logger.Fields{
			"error": scanErr.Error(),
			"value": val.Val(),
		})

		return 0, time.Time{}, scanErr
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

	return tokens, lastUpdate, nil
}

func (rl *RateLimiter) isRateLimited(tokens int, lastUpdate time.Time, err error, ctx context.Context, key string) bool {
	if tokens > 0 {
		return false
	}

	value := fmt.Sprintf("%d:%d", tokens, lastUpdate.Unix())

	var setErr error

	if err != nil && err.Error() == errRedisNil {
		expiration := rl.rate
		_, setErr = rl.redis.Set(ctx, key, value, expiration)
	} else {
		_, setErr = rl.redis.SetWithTTL(ctx, key, value)
	}

	if setErr != nil {
		rl.logger.Error("Failed to update rate limit data", logger.Fields{
			"error": setErr.Error(),
			"key":   key,
		})
	}

	rl.logger.Debug(errRateLimitExceeded, logger.Fields{
		"tokens": tokens,
		"key":    key,
	})

	return true
}

func (rl *RateLimiter) consumeTokenAndPersist(tokens int, lastUpdate time.Time, err error, ctx context.Context, key string) {
	tokens--
	value := fmt.Sprintf("%d:%d", tokens, lastUpdate.Unix())

	var setErr error

	if err != nil && err.Error() == errRedisNil {
		expiration := rl.rate
		_, setErr = rl.redis.Set(ctx, key, value, expiration)
	} else {
		_, setErr = rl.redis.SetWithTTL(ctx, key, value)
	}

	if setErr != nil {
		rl.logger.Error("Failed to update rate limit data", logger.Fields{
			"error": setErr.Error(),
			"key":   key,
		})

		return
	}

	rl.logger.Debug("Rate limit updated", logger.Fields{
		"remainingTokens": tokens,
		"key":             key,
	})
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := getClientID(c)

		if t, found := rl.recentOffenders.Load(clientID); found {
			if time.Since(t.(time.Time)) < 2*time.Second {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": errRateLimitExceeded,
				})

				c.Abort()
				return
			}
		}

		if !rl.Allow(clientID) {
			rl.recentOffenders.Store(clientID, rl.timeNow())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": errRateLimitExceeded,
			})

			c.Abort()

			return
		}

		c.Next()
	}
}

func RateLimit(capacity int, rate time.Duration, isRedisEnabled bool) gin.HandlerFunc {
	if !isRedisEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter, err := NewRateLimiter(capacity, rate)

	if err != nil {
		logger.New(config.GetLogLevel(), os.Stdout).Error("Failed to create rate limiter", logger.Fields{"error": err.Error()})

		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: rate limiter unavailable"})
			c.Abort()
		}
	}

	return func(c *gin.Context) {
		clientID := getClientID(c)

		if !limiter.Allow(clientID) {
			limiter.logger.Warn(errRateLimitExceeded, logger.Fields{
				"client_id": clientID,
				"path":      c.Request.URL.Path,
			})

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": errRateLimitExceeded,
			})

			c.Abort()
			return
		}

		c.Next()
	}
}
