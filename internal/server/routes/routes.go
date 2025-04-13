package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router gin.IRouter) {
	router.GET("/", Index)
	router.GET("/health", HealthCheck)

	router.GET("/login", Login)
}
