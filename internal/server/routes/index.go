package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type IndexData struct {
	Title       string
	Description string
}

func Index(c *gin.Context) {
	data := IndexData{
		Title:       "INDEX",
		Description: "INDEX",
	}

	c.HTML(http.StatusOK, "index.html.tmpl", data)
}
