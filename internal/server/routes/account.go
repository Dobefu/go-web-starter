package routes

import (
	"github.com/gin-gonic/gin"
)

func Account(c *gin.Context) {
	data := RouteData{
		Template:    "pages/account",
		Title:       "Account",
		Description: "View and manage your account details.",
		HttpStatus:  200,
	}

	RenderRouteHTML(c, data)
}
