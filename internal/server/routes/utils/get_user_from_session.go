package utils

import (
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var userFindByID = user.FindByID

func GetUserFromSession(c *gin.Context) *user.User {
	session := sessions.Default(c)
	userID := session.Get("userID")

	if userID == nil {
		return nil
	}

	dbVal, exists := c.Get("db")

	if !exists {
		return nil
	}

	db, ok := dbVal.(database.DatabaseInterface)

	if !ok {
		return nil
	}

	id, ok := userID.(int)

	if !ok {
		return nil
	}

	currentUser, err := userFindByID(db, id)

	if err != nil {
		log := logger.New(config.GetLogLevel(), os.Stdout)

		log.Error("getCurrentUser: failed to load user", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})

		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()

		return nil
	}

	return currentUser
}
