package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFound(c *gin.Context) {
	data := RouteData{
		Template:   "pages/not-found",
		HttpStatus: http.StatusNotFound,

		Title:       "Page Not Found",
		Description: "The page you are looking for does not exist",
	}

	RenderRouteHTML(c, data)
}
