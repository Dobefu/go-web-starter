package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	data := RouteData{
		Title:       "INDEX",
		Description: "INDEX",
		Template:    "pages/index",
	}

	c.HTML(http.StatusOK, data.Template, data)
}
