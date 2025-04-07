package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RenderRouteHTML(c *gin.Context, routeData RouteData) {
	data := struct {
		RouteData
		SiteName string
		Year     string
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
		Year:      time.Now().Format("2006"),
	}

	c.HTML(data.HttpStatus, data.Template, data)
}
