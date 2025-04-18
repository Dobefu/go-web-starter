package routes

import (
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	data := RouteData{
		Template:    "pages/login",
		HttpStatus:  http.StatusOK,
		Title:       "Log In",
		Description: "Sign in to your account",
		FormData: FormData{
			Values:   v.GetFormData(),
			Errors:   v.GetSessionErrors(),
			Messages: v.GetMessages(),
		},
	}

	v.ClearSession()
	RenderRouteHTML(c, data)
}

func LoginPost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	err := v.ValidateForm(c.Request)

	if err != nil {
		log.Error("Failed to parse form data", map[string]any{
			"error": err.Error(),
		})
	}

	email := v.GetFormValue(c.Request, "email")
	password := v.GetFormValue(c.Request, "password")

	v.Required("email", email)

	v.Required("password", password)
	v.MinLength("password", password, 8)

	if v.HasErrors() {
		v.SetFormData(map[string]string{
			"email":    email,
			"password": password,
		})

		v.SetErrors()
		v.SetFlash("Please correct the errors below")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	log.Info("Login attempt", map[string]any{
		"email": email,
	})

	v.SetFlash("Successfully logged in!")
	c.Redirect(http.StatusSeeOther, "/")
}
