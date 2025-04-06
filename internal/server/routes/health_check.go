package routes

import (
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	db, exists := c.Get("db")

	if !exists {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  "Database connection not found",
		})

		return
	}

	database, ok := db.(database.DatabaseInterface)

	if !ok {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  "Invalid database connection type",
		})

		return
	}

	err := database.Ping()

	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})

		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"error":  nil,
	})
}
