package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	data := RouteData{
		Title:       "INDEX",
		Description: "INDEX",
	}

	c.HTML(http.StatusOK, "index.html.tmpl", data)
}
