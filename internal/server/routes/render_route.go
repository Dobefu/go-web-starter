package routes

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RenderRouteHTML(c *gin.Context, routeData RouteData) {
	v := validator.New()
	v.SetContext(c)

	var canonical string

	if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
		canonical = fmt.Sprintf("%s%s", viper.GetString("site.host"), c.Request.URL.Path)
		canonical = strings.TrimRight(canonical, "/")
	}

	data := struct {
		RouteData
		SiteName  string
		Year      string
		Nonce     string
		BuildHash string
		Canonical string
		Messages  []message.Message
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
		Year:      time.Now().Format("2006"),
		Nonce:     c.GetString("nonce"),
		BuildHash: config.BuildHash,
		Canonical: canonical,
		Messages:  v.GetMessages(),
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
			_ = c.Error(err)
			return
		}

		_, _ = c.Writer.Write(buf.Bytes())

		return
	}

	c.HTML(data.HttpStatus, data.Template, data)
}
