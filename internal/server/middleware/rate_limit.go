package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*TokenBucket
	capacity int
	rate     time.Duration
}

type TokenBucket struct {
	tokens     int
	lastUpdate time.Time
	capacity   int
}

func NewRateLimiter(capacity int, rate time.Duration) *RateLimiter {
	return &RateLimiter{
		clients:  make(map[string]*TokenBucket),
		capacity: capacity,
		rate:     rate,
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
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.clients[clientID]

	if !exists {
		bucket = &TokenBucket{
			tokens:     rl.capacity,
			lastUpdate: time.Now(),
			capacity:   rl.capacity,
		}

		rl.clients[clientID] = bucket
	}

	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := int(elapsed / rl.rate)

	if tokensToAdd > 0 {
		bucket.tokens = min(bucket.capacity, bucket.tokens+tokensToAdd)
		bucket.lastUpdate = now
	}

	if bucket.tokens > 0 {
		bucket.tokens -= 1
		return true
	}

	return false
}

func RateLimit(capacity int, rate time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(capacity, rate)

	return func(c *gin.Context) {
		clientID := getClientIP(c)

		if !limiter.Allow(clientID) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()

			return
		}

		c.Next()
	}
}
