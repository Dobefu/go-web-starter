package routes

import (
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router gin.IRouter) {
	router.GET("/", Index)
	router.GET("/health", HealthCheck)

	router.GET("/robots.txt", RobotsTxt)

	anonOnly := router.Group("/")
	anonOnly.Use(middleware.AnonOnly())
	anonOnly.GET("/login", Login)
	anonOnly.POST("/login", LoginPost)
	anonOnly.GET("/register", Register)

	authOnly := router.Group("/")
	authOnly.Use(middleware.AuthOnly())
}
