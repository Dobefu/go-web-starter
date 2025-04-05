package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RegisterRoutes(router gin.IRouter) {
	router.GET("/", Index)
	router.GET("/health", HealthCheck)
}

func RenderRoute(c *gin.Context, routeData RouteData) {
	data := struct {
		RouteData
		SiteName string
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
	}

	c.HTML(http.StatusOK, data.Template, data)
}
