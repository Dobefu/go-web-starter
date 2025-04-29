package utils

import (
	"errors"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
)

func GetDbFromContext(c *gin.Context) (db database.DatabaseInterface, err error) {
	dbVal, exists := c.Get("db")

	if !exists {
		return nil, errors.New("database not found in context")
	}

	db, ok := dbVal.(database.DatabaseInterface)

	if !ok {
		return nil, errors.New("database in context is not of type DatabaseInterface")
	}

	return db, nil
}
