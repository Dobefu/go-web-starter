package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	data := RouteData{
		Template:   "pages/login",
		HttpStatus: http.StatusOK,

		Title:       "LOGIN",
		Description: "",
	}

	RenderRouteHTML(c, data)
}
