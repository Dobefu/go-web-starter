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

	v.ValidEmail("email", email)
	v.Required("email", email)

	v.Required("username", username)
	v.MinLength("username", username, 3)

	v.Required("password", password)
	v.MinLength("password", password, 8)
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
		route_utils.RedirectWithError(
			c,
			v,
			map[string]string{
				"username": username,
				"email":    email,
			},
			"Please correct the errors below",
			"/register",
		)
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

	token := usr.CreateVerificationToken()

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
			Template: "email/register_verify",
			Data: map[string]any{
				"Username": username,
				"Token":    token,
				"Email":    email,
			},
		},
	)

	if err != nil {
		log.Error("Failed to send the registation email", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	v.SetFlash(message.Message{
		Type: message.MessageTypeSuccess,
		Body: "Your account has been created! Please check you inbox for further instructions.",
	})

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/register/verify?email=%s", email))
}
