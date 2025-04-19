package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var csrfRandRead = rand.Read

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		session := sessions.Default(c)
		token := session.Get("csrf_token")
		formToken := c.PostForm("_csrf")

		if token == nil || formToken == "" || token.(string) != formToken {
			session := sessions.Default(c)
			session.AddFlash("Invalid or expired form submission. Please try again.")
			session.Save()

			c.Redirect(http.StatusSeeOther, c.Request.URL.Path)
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetCSRFToken(c *gin.Context) string {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	session := sessions.Default(c)
	token := session.Get("csrf_token")

	if token == nil {
		b := make([]byte, 32)
		_, err := csrfRandRead(b)

		if err != nil {
			log.Error("Failed to generate CSRF token", logger.Fields{"Error": err.Error()})
			return ""
		}

		token = base64.StdEncoding.EncodeToString(b)
		session.Set("csrf_token", token)
		session.Save()
	}

	return token.(string)
}
