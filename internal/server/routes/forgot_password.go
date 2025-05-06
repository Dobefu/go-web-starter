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
	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	msgPasswdReset = "If your email address exists and has an account associated, password reset instructions have been sent."

	errUserPasswdVerify = "Could not verify the password reset token."
)

func ForgotPassword(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	email := v.GetFormValue(c.Request, "email")
	token := v.GetFormValue(c.Request, "token")

	if len(email) > 0 && len(token) > 0 {
		log := logger.New(config.GetLogLevel(), os.Stdout)
		db, err := route_utils.GetDbFromContext(c)

		if err != nil {
			log.Error("Could not get the database from the context", logger.Fields{"err": err.Error()})
			RenderRouteHTML(c, GenericErrorData(c))

			return
		}

		usr, err := user.FindByEmail(db, email)

		if err != nil {
			log.Warn("Could not get the user from the email address", logger.Fields{"err": err.Error()})

			v.SetFlash(message.Message{
				Type: message.MessageTypeError,
				Body: errUserPasswdVerify,
			})

			c.Redirect(http.StatusSeeOther, paths.PathRegister)
			return
		}

		expectedToken := usr.CreateVerificationToken()

		if token != expectedToken || !usr.GetStatus() {
			v.SetFlash(message.Message{
				Type: message.MessageTypeError,
				Body: errUserPasswdVerify,
			})

			c.Redirect(http.StatusSeeOther, paths.PathRegister)
			return
		}

		session := getSession(c)
		err = usr.Login(db, session)

		if err != nil {
			log.Error("Failed to save session after verification login", logger.Fields{"err": err.Error()})
			RenderRouteHTML(c, GenericErrorData(c))

			return
		}

		v.SetFlash(message.Message{
			Type: message.MessageTypeSuccess,
			Body: "You have used your one-time login token. Please change your password.",
		})

		c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/edit", paths.PathAccount))
		return
	}

	csrfToken := middleware.GetCSRFToken(c)

	data := RouteData{
		Template:   "pages/forgot_password",
		HttpStatus: http.StatusOK,

		Title:       "Reset your password",
		Description: "Request a new password.",

		FormData: FormData{
			Values: v.GetFormData(),
			Errors: v.GetSessionErrors(),
		},
		CSRFToken: csrfToken,
	}

	RenderRouteHTML(c, data)

	v.ClearSession()
}

func ForgotPasswordPost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	err := v.ValidateForm(c.Request)

	if err != nil {
		log.Error("Failed to parse form data", map[string]any{"error": err.Error()})
	}

	email := v.GetFormValue(c.Request, "email")

	v.ValidEmail("email", email)
	v.Required("email", email)

	if v.HasErrors() {
		v.SetFormData(map[string]string{
			"email": email,
		})

		v.SetErrors()
		v.SetFlash(message.Message{Body: "Please correct the errors below"})
		c.Redirect(http.StatusSeeOther, paths.PathForgotPassword)

		return
	}

	db, err := route_utils.GetDbFromContext(c)

	if err != nil {
		log.Error("Failed to get database connection from context", nil)
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	foundUser, err := findByEmail(db, email)

	if err != nil || !foundUser.GetStatus() {
		v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: msgPasswdReset})
		c.Redirect(http.StatusSeeOther, paths.PathLogin)

		return
	}

	log.Info("Password reset request successful", map[string]any{
		"email":  email,
		"userID": foundUser.GetID(),
	})

	token := foundUser.CreateVerificationToken()

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
		fmt.Sprintf("Reset your %s password", viper.GetString("site.name")),
		emailer.EmailBody{
			Template: "email/forgot_password",
			Data: map[string]any{
				"Username": foundUser.GetUsername(),
				"Token":    token,
				"Email":    email,
			},
		},
	)

	if err != nil {
		log.Error("Failed to send the password reset email", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: msgPasswdReset})
	c.Redirect(http.StatusSeeOther, paths.PathLogin)
}
