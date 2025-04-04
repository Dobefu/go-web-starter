package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RenderRouteHTML(c *gin.Context, routeData RouteData) {
	data := struct {
		RouteData
		SiteName string
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
	}

	c.HTML(data.HttpStatus, data.Template, data)
}
