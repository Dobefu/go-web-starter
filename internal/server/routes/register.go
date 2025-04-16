package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	data := RouteData{
		Template:   "pages/register",
		HttpStatus: http.StatusOK,

		Title:       "Register",
		Description: "Register a new account",
	}

	RenderRouteHTML(c, data)
}
