package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	data := RouteData{
		Template:   "pages/login",
		HttpStatus: http.StatusOK,

		Title:       "Log In",
		Description: "Sign in to your account",
	}

	RenderRouteHTML(c, data)
}
