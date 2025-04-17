package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router gin.IRouter) {
	router.GET("/", Index)
	router.GET("/health", HealthCheck)

	router.GET("/robots.txt", RobotsTxt)

	router.GET("/login", Login)
	router.GET("/register", Register)
}
