package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

func DynamicContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() == gin.DebugMode {
			c.Next()
			return
		}

		buf := new(bytes.Buffer)
		originalWriter := c.Writer

		if rw, ok := originalWriter.(*ResponseWriter); ok {
			c.Next()

			content := rw.body.Bytes()
			contentStr := string(content)

			writeResponse(c, []byte(contentStr), "PROCESSED")
			return
		}

		c.Writer = &ResponseWriter{
			ResponseWriter: originalWriter,
			body:           buf,
		}

		c.Next()

		content := buf.Bytes()
		contentStr := string(content)

		writeResponse(c, []byte(contentStr), "PROCESSED")
	}
}
