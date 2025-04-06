package middleware

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type RedisMiddlewareMock struct {
	redis.RedisInterface
}

func TestRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRedis := &RedisMiddlewareMock{}

	c, _ := gin.CreateTestContext(nil)

	middleware := Redis(mockRedis)
	middleware(c)

	redis, exists := c.Get("redis")
	assert.True(t, exists)
	assert.Equal(t, mockRedis, redis)
}
