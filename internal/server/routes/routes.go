package routes

import (
	"github.com/gin-gonic/gin"
)

// Register sets up all the routes for the application
func Register(router gin.IRouter) {
	router.GET("/health", healthCheck)
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
