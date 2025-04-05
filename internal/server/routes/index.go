package routes

import (
	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	data := RouteData{
		Title:       "INDEX",
		Description: "INDEX",
		Template:    "pages/index",
	}

	RenderRoute(c, data)
}
