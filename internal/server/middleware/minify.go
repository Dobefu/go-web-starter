package middleware

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/json"
)

const contentTypeHeader = "Content-Type"

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

		if gin.Mode() == gin.DebugMode {
			processRequest(c, m, log)
			return
		}

		buf := new(bytes.Buffer)
		originalWriter := c.Writer

		c.Writer = &ResponseWriter{
			ResponseWriter: originalWriter,
			body:           buf,
		}

		c.Next()

		contentType := originalWriter.Header().Get(contentTypeHeader)
		content := buf.Bytes()

		if bytes.Contains(content, []byte("dynamic-content")) {
			writeResponse(c, content, "SKIP")
			return
		}

		cacheKey := fmt.Sprintf("%s:%s:%s", config.BuildHash, c.Request.Method, c.Request.URL.Path)
		cache := GetMinifyCache()

		if cachedContent := cache.Get(cacheKey); cachedContent != nil {
			c.Writer.Header().Set(contentTypeHeader, "text/html")
			writeResponse(c, cachedContent, "HIT")
			c.Abort()

			return
		}

		minifiedBytes := processResponse(c, m, buf, contentType, log)
		cache.Set(cacheKey, minifiedBytes, time.Hour)
		writeResponse(c, minifiedBytes, "MISS")
	}
}

func processRequest(c *gin.Context, m *minify.M, log *logger.Logger) {
	buf := new(bytes.Buffer)
	originalWriter := c.Writer

	c.Writer = &ResponseWriter{
		ResponseWriter: originalWriter,
		body:           buf,
	}

	c.Next()

	contentType := originalWriter.Header().Get(contentTypeHeader)
	minifiedBytes := processResponse(c, m, buf, contentType, log)

	writeResponse(c, minifiedBytes, "MISS")
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

func writeResponse(c *gin.Context, content []byte, cacheStatus string) {
	var originalWriter gin.ResponseWriter
	rw, ok := c.Writer.(*ResponseWriter)

	if ok {
		originalWriter = rw.ResponseWriter
	} else {
		originalWriter = c.Writer
	}

	originalWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	originalWriter.Header().Set("X-Cache", cacheStatus)

	_, _ = originalWriter.Write(content)
}
