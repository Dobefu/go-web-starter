package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	data := RouteData{
		Template:   "pages/register",
		HttpStatus: http.StatusOK,

		Title:       "REGISTER",
		Description: "",
	}

	RenderRouteHTML(c, data)
}
