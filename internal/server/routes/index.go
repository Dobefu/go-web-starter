package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	data := RouteData{
		Template:   "pages/index",
		HttpStatus: http.StatusOK,

		Title:       "INDEX",
		Description: "INDEX",
	}

	RenderRouteHTML(c, data)
}
