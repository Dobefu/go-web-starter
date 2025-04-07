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

		if cachedContent := getCachedContent(c, log); cachedContent != nil {
			writeResponse(c, cachedContent, "HIT")
			return
		}

		buf := new(bytes.Buffer)
		originalWriter := c.Writer
		c.Writer = &ResponseWriter{
			ResponseWriter: originalWriter,
			body:           buf,
		}

		c.Next()

		contentType := originalWriter.Header().Get("Content-Type")
		minifiedBytes := processResponse(c, m, buf, contentType, log)

		cacheMinifiedContent(c, minifiedBytes, log)

		writeResponse(c, minifiedBytes, "MISS")
	}
}

func getCachedContent(c *gin.Context, log *logger.Logger) []byte {
	redisClient, exists := c.Get("redis")

	if !exists {
		return nil
	}

	redis := redisClient.(redis.RedisInterface)
	cacheKey := fmt.Sprintf("minify:%s:%s", c.Request.Method, c.Request.URL.Path)
	ctx := context.Background()

	cachedCmd, err := redis.Get(ctx, cacheKey)

	if err != nil || cachedCmd == nil {
		return nil
	}

	cachedContent := cachedCmd.Val()
	if cachedContent == "" {
		return nil
	}

	cachedBytes := []byte(cachedContent)

	log.Trace("Using cached minified content", logger.Fields{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"key":    cacheKey,
		"size":   len(cachedBytes),
	})

	return cachedBytes
}

func processResponse(c *gin.Context, m *minify.M, buf *bytes.Buffer, contentType string, log *logger.Logger) []byte {
	// If there's no corresponding minify function, return the original data.
	_, _, minifierFunc := m.Match(contentType)
	if minifierFunc == nil {
		log.Trace("No minifier found for content type", logger.Fields{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"contentType": contentType,
		})
		return buf.Bytes()
	}

	minified, err := m.String(contentType, buf.String())
	if err != nil {
		log.Trace("Minification failed, using original content", logger.Fields{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"contentType": contentType,
			"error":       err.Error(),
		})

		return buf.Bytes()
	}

	return []byte(minified)
}

func cacheMinifiedContent(c *gin.Context, minifiedBytes []byte, log *logger.Logger) {
	redisClient, exists := c.Get("redis")
	if !exists {
		return
	}

	redis := redisClient.(redis.RedisInterface)
	cacheKey := fmt.Sprintf("minify:%s:%s", c.Request.Method, c.Request.URL.Path)
	ctx := context.Background()

	_, err := redis.Set(ctx, cacheKey, minifiedBytes, time.Hour)

	if err != nil {
		log.Trace("Failed to cache minified content", logger.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"key":    cacheKey,
			"error":  err.Error(),
		})

		return
	}

	log.Trace("Cached minified content", logger.Fields{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"key":    cacheKey,
		"size":   len(minifiedBytes),
	})
}

func writeResponse(c *gin.Context, content []byte, cacheStatus string) {
	var originalWriter gin.ResponseWriter

	if rw, ok := c.Writer.(*ResponseWriter); ok {
		originalWriter = rw.ResponseWriter
	} else {
		originalWriter = c.Writer
	}

	originalWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	originalWriter.Header().Set("X-Cache", cacheStatus)

	_, _ = originalWriter.Write(content)
}
