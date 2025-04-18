package middleware

import (
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetFlashMessage(c *gin.Context) string {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	session := sessions.Default(c)
	flashes := session.Flashes()

	if len(flashes) == 0 {
		return ""
	}

	msg, ok := flashes[0].(string)

	if !ok {
		return ""
	}

	err := session.Save()

	if err != nil {
		log.Error("Failed to save session after clearing flashes", logger.Fields{
			"error": err.Error(),
		})
	}

	return msg
}

func Flash() gin.HandlerFunc {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	return func(c *gin.Context) {
		v := validator.New()
		v.SetContext(c)

		c.Set("AddFlash", func(message string) {
			session := sessions.Default(c)
			session.AddFlash(message)
			err := session.Save()

			if err != nil {
				log.Error("Failed to save session after adding flash", logger.Fields{
					"error": err.Error(),
				})
			}
		})

		c.Set("GetFlash", func() string {
			session := sessions.Default(c)
			flashes := session.Flashes()

			if len(flashes) == 0 {
				return ""
			}

			msg, ok := flashes[0].(string)

			if !ok {
				return ""
			}

			err := session.Save()

			if err != nil {
				log.Error("Failed to save session after clearing flashes", logger.Fields{
					"error": err.Error(),
				})
			}

			return msg
		})

		c.Next()
	}
}
