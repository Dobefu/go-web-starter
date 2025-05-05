package routes

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/message"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/Dobefu/go-web-starter/internal/user"
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

	currentUser := route_utils.GetUserFromSession(c)

	data := struct {
		RouteData
		SiteName  string
		Year      string
		Nonce     string
		BuildHash string
		Canonical string
		Href      string
		Messages  []message.Message
		User      *user.User
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
		Year:      time.Now().Format("2006"),
		Nonce:     c.GetString("nonce"),
		BuildHash: config.BuildHash,
		Canonical: canonical,
		Href:      c.Request.URL.Path,
		Messages:  v.GetMessages(),
		User:      currentUser,
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
