package routes

import (
	"github.com/gin-gonic/gin"
)

func Account(c *gin.Context) {
	data := RouteData{
		Template:    "pages/account",
		Title:       "My Account",
		Description: "View your account details.",
		HttpStatus:  200,
	}

	RenderRouteHTML(c, data)
}
