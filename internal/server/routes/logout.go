package routes

import (
	"net/http"

	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()

	v := validator.New()
	v.SetContext(c)
	v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: "You have been logged out"})

	c.Redirect(http.StatusSeeOther, paths.PathLogin)
}
