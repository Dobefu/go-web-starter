package routes

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	router.GET("/health", HealthCheck)
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
