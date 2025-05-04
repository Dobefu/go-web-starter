package routes

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var userFindByID = user.FindByID

func getCurrentUser(c *gin.Context) *user.User {
	session := sessions.Default(c)
	userID := session.Get("userID")

	if userID == nil {
		return nil
	}

	dbVal, exists := c.Get("db")

	if !exists {
		return nil
	}

	db, ok := dbVal.(database.DatabaseInterface)

	if !ok {
		return nil
	}

	id, ok := userID.(int)

	if !ok {
		return nil
	}

	currentUser, err := userFindByID(db, id)

	if err != nil {
		log := logger.New(config.GetLogLevel(), os.Stdout)

		log.Error("getCurrentUser: failed to load user", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})

		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()

		return nil
	}

	return currentUser
}

func RenderRouteHTML(c *gin.Context, routeData RouteData) {
	v := validator.New()
	v.SetContext(c)

	var canonical string

	if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
		canonical = fmt.Sprintf("%s%s", viper.GetString("site.host"), c.Request.URL.Path)
		canonical = strings.TrimRight(canonical, "/")
	}

	currentUser := getCurrentUser(c)

	data := struct {
		RouteData
		SiteName  string
		Year      string
		Nonce     string
		BuildHash string
		Canonical string
		Messages  []message.Message
		User      *user.User
	}{
		RouteData: routeData,
		SiteName:  viper.GetString("site.name"),
		Year:      time.Now().Format("2006"),
		Nonce:     c.GetString("nonce"),
		BuildHash: config.BuildHash,
		Canonical: canonical,
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
