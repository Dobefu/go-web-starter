package utils

import (
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
)

func GetDbFromContext(c *gin.Context) (db database.DatabaseInterface, err error) {
	dbVal, exists := c.Get("db")

	if !exists {
		return nil, err
	}

	db, ok := dbVal.(database.DatabaseInterface)

	if !ok {

		return nil, err
	}

	return db, nil
}
