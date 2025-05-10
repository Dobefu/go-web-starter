package middleware

import (
	"net/http"

	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		if session.Get("userID") == nil {
			c.Redirect(http.StatusSeeOther, paths.PathLogin)
			c.Abort()

			return
		}

		c.Next()
	}
}

func AnonOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		if session.Get("userID") != nil {
			c.Redirect(http.StatusSeeOther, paths.PathAccount)
			c.Abort()

			return
		}

		c.Next()
	}
}
