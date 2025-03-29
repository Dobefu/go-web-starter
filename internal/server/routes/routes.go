package routes

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	router.GET("/health", HealthCheck)
}
