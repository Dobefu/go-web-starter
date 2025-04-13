package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	data := RouteData{
		Template:   "pages/index",
		HttpStatus: http.StatusOK,

		Title:       "Welcome to Go Web Starter",
		Description: "A modern, production-ready Go web application template with best practices and common features pre-configured",
	}

	RenderRouteHTML(c, data)
}
