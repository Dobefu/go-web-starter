package middleware

import (
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-gonic/gin"
)

func Redis(redis redis.RedisInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("redis", redis)
		c.Next()
	}
}
