package middleware

import (
	"bytes"
	"fmt"

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
			_, _ = originalWriter.Write(buf.Bytes())
			return
		}

		minified, err := m.String(contentType, buf.String())

		if err != nil {
			c.Data(originalWriter.Status(), contentType, buf.Bytes())
			return
		}

		minifiedBytes := []byte(minified)
		originalWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(minifiedBytes)))
		_, _ = originalWriter.Write(minifiedBytes)
	}
}
