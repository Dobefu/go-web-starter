package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	data := RouteData{
		Template:   "pages/login",
		HttpStatus: http.StatusOK,

		Title:       "Log In",
		Description: "Sign in to your account",
	}

	RenderRouteHTML(c, data)
}

func LoginPost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	err := c.Request.ParseForm()

	if err != nil {
		log.Error(err.Error(), nil)
		c.Data(500, "text/plain", nil)
		return
	}

	fmt.Println(c.Request.FormValue("email"))
}
