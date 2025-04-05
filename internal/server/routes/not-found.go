package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFound(c *gin.Context) {
	data := RouteData{
		Title:       "Not found",
		Description: "The page could not be found",
		Template:    "pages/not-found",
	}

	c.HTML(http.StatusOK, data.Template, data)
}
