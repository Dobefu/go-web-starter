package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HomeData struct {
	Title   string
	Content string
}

func Index(c *gin.Context) {
	data := HomeData{
		Title: "INDEX",
	}

	c.HTML(http.StatusOK, "index.html", data)
}
