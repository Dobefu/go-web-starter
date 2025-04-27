package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		if session.Get("userID") == nil {
			c.Redirect(http.StatusSeeOther, "/login")
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
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()

			return
		}

		c.Next()
	}
}
