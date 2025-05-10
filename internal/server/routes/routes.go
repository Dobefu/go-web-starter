package routes

import (
	"fmt"

	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router gin.IRouter) {
	router.GET("/", Index)
	router.GET("/health", HealthCheck)

	router.GET("/robots.txt", RobotsTxt)

	RegisterAnonOnlyRoutes(router.Group("/"))
	RegisterAuthOnlyRoutes(router.Group("/"))
}

func RegisterAnonOnlyRoutes(rg *gin.RouterGroup) {
	rg.Use(middleware.AnonOnly())

	rg.GET(paths.PathLogin, Login)
	rg.POST(paths.PathLogin, LoginPost)
	rg.GET(paths.PathRegister, Register)
	rg.POST(paths.PathRegister, RegisterPost)
	rg.GET(fmt.Sprintf("%s/verify", paths.PathRegister), RegisterVerify)
	rg.GET(paths.PathForgotPassword, ForgotPassword)
	rg.POST(paths.PathForgotPassword, ForgotPasswordPost)
}

func RegisterAuthOnlyRoutes(rg *gin.RouterGroup) {
	rg.Use(middleware.AuthOnly())

	rg.GET(paths.PathLogout, Logout)
	rg.GET(paths.PathAccount, Account)
	rg.GET(fmt.Sprintf("%s/edit", paths.PathAccount), AccountEdit)
	rg.POST(fmt.Sprintf("%s/edit", paths.PathAccount), AccountEditPost)
	rg.GET(fmt.Sprintf("%s/delete", paths.PathAccount), AccountDelete)
	rg.POST(fmt.Sprintf("%s/delete", paths.PathAccount), AccountDeletePost)
}
