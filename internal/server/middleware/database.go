package middleware

import (
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
)

func Database(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}
