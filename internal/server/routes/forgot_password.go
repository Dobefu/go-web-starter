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
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const msgPasswdReset = "If your email address exists and has an account associated, password reset instructions have been sent."

func ForgotPassword(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

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
		c.Redirect(http.StatusSeeOther, "/forgot-password")

		return
	}

	db, err := route_utils.GetDbFromContext(c)

	if err != nil {
		log.Error("Failed to get database connection from context", nil)
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	foundUser, err := findByEmail(db, email)

	if err != nil {
		v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: msgPasswdReset})
		c.Redirect(http.StatusSeeOther, "/login")

		return
	}

	if !foundUser.GetStatus() {
		v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: msgPasswdReset})
		c.Redirect(http.StatusSeeOther, "/login")

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
	c.Redirect(http.StatusSeeOther, "/login")
}
