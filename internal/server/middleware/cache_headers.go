package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const cacheDuration = time.Hour * 24 * 7

func CacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		if isStaticAsset(c.Request.URL.Path) {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cacheDuration.Seconds())))
			c.Header("Expires", time.Now().Add(cacheDuration).Format(time.RFC1123))
		} else {
			c.Header("Cache-Control", "no-cache, must-revalidate")
			c.Header("Pragma", "no-cache")
		}

		c.Next()
	}
}

func isStaticAsset(path string) bool {
	staticExtensions := []string{
		".css",
		".js",
		".jpg",
		".jpeg",
		".png",
		".gif",
		".ico",
		".svg",
		".woff",
		".woff2",
		".ttf",
		".eot",
	}

	for _, ext := range staticExtensions {
		if len(path) < len(ext) {
			continue
		}

		if path[len(path)-len(ext):] == ext {
			return true
		}
	}

	return false
}
