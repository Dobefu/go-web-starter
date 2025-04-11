package routes

import (
	"bytes"
	"time"

	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RenderRouteHTML(c *gin.Context, routeData RouteData) {
	data := struct {
		RouteData
		SiteName string
		Year     string
		Nonce    string
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
		Year:      time.Now().Format("2006"),
		Nonce:     c.GetString("nonce"),
	}

	if gin.Mode() == gin.DebugMode {
		c.HTML(data.HttpStatus, data.Template, data)
		return
	}

	cache := templates.GetTemplateCache()
	tmpl, ok := cache.Get(routeData.Template)

	if ok && tmpl != nil {
		c.Status(data.HttpStatus)

		buf := new(bytes.Buffer)

		if err := tmpl.Execute(buf, data); err != nil {
			c.Error(err)
			return
		}

		c.Writer.Write(buf.Bytes())
		return
	}

	c.HTML(data.HttpStatus, data.Template, data)
}
