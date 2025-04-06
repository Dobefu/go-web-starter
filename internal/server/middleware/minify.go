package middleware

import (
	"bytes"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func Minify() gin.HandlerFunc {
	textHtmlMime := "text/html"

	m := minify.New()
	m.Add(textHtmlMime, &html.Minifier{
		KeepDocumentTags: true,
	})

	return func(c *gin.Context) {
		buf := new(bytes.Buffer)

		originalWriter := c.Writer

		c.Writer = &ResponseWriter{
			ResponseWriter: originalWriter,
			body:           buf,
		}

		c.Next()

		contentType := originalWriter.Header().Get("Content-Type")
		if contentType == textHtmlMime || contentType == "text/html; charset=utf-8" {
			minified, err := m.String(textHtmlMime, buf.String())

			if err != nil {
				c.Data(originalWriter.Status(), contentType, buf.Bytes())
				return
			}

			minifiedBytes := []byte(minified)
			originalWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(minifiedBytes)))
			_, _ = originalWriter.Write(minifiedBytes)
		} else {
			_, _ = originalWriter.Write(buf.Bytes())
		}
	}
}
