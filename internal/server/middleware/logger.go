package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.New(config.GetLogLevel(), os.Stdout)

		// Get the delta time between the start and the end of the request.
		startTime := time.Now()
		c.Next()
		stopTime := time.Now()

		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		if rawQuery != "" {
			path = fmt.Sprintf("%s?%s", path, rawQuery)
		}

		log.Info(
			"Gin request",
			logger.Fields{
				"status":    fmt.Sprintf("%d", c.Writer.Status()),
				"time":      fmt.Sprintf("%v", stopTime.Sub(startTime)),
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      path,
			},
		)
	}
}
