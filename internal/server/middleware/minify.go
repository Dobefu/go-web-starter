package middleware

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/json"
)

type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func Minify() gin.HandlerFunc {
	m := minify.New()
	m.Add("text/html", &html.Minifier{KeepDocumentTags: true})
	m.Add("application/json", &json.Minifier{})

	return func(c *gin.Context) {
		log := logger.New(config.GetLogLevel(), os.Stdout)

		if c.Request.Method != "GET" {
			log.Trace("Skipping minification for non-GET request", logger.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			})
			c.Next()
			return
		}

		redisClient, exists := c.Get("redis")

		if exists {
			redis := redisClient.(redis.RedisInterface)
			cacheKey := fmt.Sprintf("minify:%s:%s", c.Request.Method, c.Request.URL.Path)
			ctx := context.Background()

			cachedCmd, err := redis.Get(ctx, cacheKey)

			if err == nil && cachedCmd != nil {
				cachedContent := cachedCmd.Val()

				if cachedContent != "" {
					cachedBytes := []byte(cachedContent)

					log.Trace("Using cached minified content", logger.Fields{
						"method": c.Request.Method,
						"path":   c.Request.URL.Path,
						"key":    cacheKey,
						"size":   len(cachedBytes),
					})

					c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(cachedBytes)))
					c.Writer.Header().Set("X-Cache", "HIT")
					_, _ = c.Writer.Write(cachedBytes)

					return
				}
			}
		}

		buf := new(bytes.Buffer)
		originalWriter := c.Writer

		c.Writer = &ResponseWriter{
			ResponseWriter: originalWriter,
			body:           buf,
		}

		c.Next()

		contentType := originalWriter.Header().Get("Content-Type")
		_, _, minifierFunc := m.Match(contentType)

		// If there's no corresponding minify function, return the original data.
		if minifierFunc == nil {
			log.Trace("No minifier found for content type", logger.Fields{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"contentType": contentType,
			})

			_, _ = originalWriter.Write(buf.Bytes())
			return
		}

		minified, err := m.String(contentType, buf.String())
		if err != nil {
			log.Trace("Minification failed, using original content", logger.Fields{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"contentType": contentType,
				"error":       err.Error(),
			})

			_, _ = originalWriter.Write(buf.Bytes())
			return
		}

		minifiedBytes := []byte(minified)

		// Cache the minified content if Redis is available.
		if redisClient, exists := c.Get("redis"); exists {
			redis := redisClient.(redis.RedisInterface)
			cacheKey := fmt.Sprintf("minify:%s:%s", c.Request.Method, c.Request.URL.Path)
			ctx := context.Background()

			_, err = redis.Set(ctx, cacheKey, minifiedBytes, time.Hour)

			if err != nil {
				log.Trace("Failed to cache minified content", logger.Fields{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"key":    cacheKey,
					"error":  err.Error(),
				})
			} else {
				log.Trace("Cached minified content", logger.Fields{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"key":    cacheKey,
					"size":   len(minifiedBytes),
				})
			}
		}

		originalWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(minifiedBytes)))
		originalWriter.Header().Set("X-Cache", "MISS")
		_, _ = originalWriter.Write(minifiedBytes)
	}
}
