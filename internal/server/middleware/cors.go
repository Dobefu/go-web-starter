package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsHeaders() gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowOrigins:     []string{"http://localhost:*"},
			AllowMethods:     []string{"HEAD", "GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowWebSockets:  true,
			AllowWildcard:    true,
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
	)
}
