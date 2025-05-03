package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	emailer "github.com/Dobefu/go-web-starter/internal/email"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var findByEmail = user.FindByEmail
var findByUsername = user.FindByUsername
var getSession = sessions.Default

func Register(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	csrfToken := middleware.GetCSRFToken(c)

	data := RouteData{
		Template:   "pages/register",
		HttpStatus: http.StatusOK,

		Title:       "Register",
		Description: "Register a new account",

		FormData: FormData{
			Values: v.GetFormData(),
			Errors: v.GetSessionErrors(),
		},
		CSRFToken: csrfToken,
	}

	RenderRouteHTML(c, data)
}

func RegisterPost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	err := v.ValidateForm(c.Request)

	if err != nil {
		log.Error("Failed to parse form data", map[string]any{"error": err.Error()})
	}

	email := v.GetFormValue(c.Request, "email")
	username := v.GetFormValue(c.Request, "username")
	password := v.GetFormValue(c.Request, "password")
	passwordConfirm := v.GetFormValue(c.Request, "password_confirm")

	v.Required("email", email)
	v.Required("username", username)
	v.Required("password", password)
	v.Required("password_confirm", passwordConfirm)
	v.PasswordsMatch("password", password, passwordConfirm)

	db, err := route_utils.GetDbFromContext(c)

	if err != nil {
		log.Error("Failed to get database connection from context", nil)
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	_, err = findByUsername(db, username)

	if err == nil {
		v.AddFieldError("username", "This username is already taken")
	}

	_, err = findByEmail(db, email)

	if err == nil {
		v.AddFieldError("email", "This email address is already taken")
	}

	if v.HasErrors() {
		redirectToRegisterWithError(c, v, username, email, "Please correct the errors below")
		return
	}

	hashedPassword, err := user.HashPassword(password)

	if err != nil {
		log.Error("Failed to save the user", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	usr := user.NewUser(username, email, hashedPassword, false)
	err = usr.Save(db)

	if err != nil {
		log.Error("Failed to save the user", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	mail := emailer.New(
		viper.GetString("email.host"),
		viper.GetString("email.port"),
		viper.GetString("email.identity"),
		viper.GetString("email.user"),
		viper.GetString("email.password"),
	)

	err = mail.SendMail(
		viper.GetString("site.email"),
		[]string{email},
		fmt.Sprintf("Activate your %s account", viper.GetString("site.name")),
		emailer.EmailBody{
			Template: "email/test_email",
		},
	)

	if err != nil {
		log.Error("Failed to send the registation email", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	v.SetFlash(message.Message{
		Type: message.MessageTypeSuccess,
		Body: "Your account has been created!",
	})

	c.Redirect(http.StatusSeeOther, "/login")
}

func redirectToRegisterWithError(c *gin.Context, v *validator.Validator, username string, email string, flashMsg string) {
	v.SetFormData(map[string]string{
		"username": username,
		"email":    email,
	})

	v.SetErrors()
	v.SetFlash(message.Message{Body: flashMsg})
	c.Redirect(http.StatusSeeOther, "/register")
}
