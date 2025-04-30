package utils

import (
	"errors"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
)

var (
	errDbNotFound  = errors.New("database not found in context")
	errDbWrongType = errors.New("database in context is not of type DatabaseInterface")
)

func GetDbFromContext(c *gin.Context) (db database.DatabaseInterface, err error) {
	dbVal, exists := c.Get("db")

	if !exists {
		return nil, errDbNotFound
	}

	db, ok := dbVal.(database.DatabaseInterface)

	if !ok {
		return nil, errDbWrongType
	}

	return db, nil
}
