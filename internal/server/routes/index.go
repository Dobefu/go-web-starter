package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type IndexData struct {
	Title string
}

func Index(c *gin.Context) {
	data := IndexData{
		Title: "INDEX",
	}

	c.HTML(http.StatusOK, "index.html.tmpl", data)
}
