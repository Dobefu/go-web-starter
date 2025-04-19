package routes

import (
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	log.Trace("Health check request received", nil)

	db, exists := c.Get("db")

	if !exists {
		log.Error("Database connection not found in context", nil)

		c.JSON(500, gin.H{
			"status": "error",
			"error":  "Database connection not found",
		})

		return
	}

	database, ok := db.(database.DatabaseInterface)

	if !ok {
		log.Error("Invalid database connection type in context", nil)

		c.JSON(500, gin.H{
			"status": "error",
			"error":  "Invalid database connection type",
		})

		return
	}

	err := database.Ping()

	if err != nil {
		log.Error("Database ping failed", logger.Fields{"error": err.Error()})

		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})

		return
	}

	log.Debug("Health check completed successfully", nil)

	c.JSON(200, gin.H{
		"status": "ok",
		"error":  nil,
	})
}
