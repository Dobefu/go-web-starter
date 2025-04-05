package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFound(c *gin.Context) {
	data := RouteData{
		Template:   "pages/not-found",
		HttpStatus: http.StatusNotFound,

		Title:       "Not found",
		Description: "The page could not be found",
	}

	RenderRouteHTML(c, data)
}
