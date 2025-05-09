package routes

import (
	"errors"
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	csrfToken := middleware.GetCSRFToken(c)

	data := RouteData{
		Template:   "pages/login",
		HttpStatus: http.StatusOK,

		Title:       "Log In",
		Description: "Sign in to your account",

		FormData: FormData{
			Values: v.GetFormData(),
			Errors: v.GetSessionErrors(),
		},
		CSRFToken: csrfToken,
	}

	RenderRouteHTML(c, data)

	v.ClearSession()
}

func LoginPost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	err := v.ValidateForm(c.Request)

	if err != nil {
		log.Error("Failed to parse form data", map[string]any{"error": err.Error()})
	}

	email := v.GetFormValue(c.Request, "email")
	password := v.GetFormValue(c.Request, "password")

	v.ValidEmail("email", email)
	v.Required("email", email)
	v.Required("password", password)

	if v.HasErrors() {
		route_utils.RedirectWithError(
			c,
			v,
			map[string]string{
				"email": email,
			},
			"Please correct the errors below",
			paths.PathLogin,
		)
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
		if errors.Is(err, user.ErrInvalidCredentials) {
			v.AddFieldError("email", user.ErrInvalidCredentials.Error())

			log.Warn("Login failed: invalid credentials (email not found)", map[string]any{"email": email})
			route_utils.RedirectWithError(
				c,
				v,
				map[string]string{
					"email": email,
				},
				user.ErrInvalidCredentials.Error(),
				paths.PathLogin,
			)
		} else {
			log.Error("Database error during login", map[string]any{"email": email, "error": err.Error()})
			RenderRouteHTML(c, GenericErrorData(c))
		}

		return
	}

	err = foundUser.CheckPassword(password)

	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			v.AddFieldError("email", user.ErrInvalidCredentials.Error())

			log.Warn("Login failed: invalid credentials (password mismatch)", map[string]any{"email": email})
			route_utils.RedirectWithError(
				c,
				v,
				map[string]string{
					"email": email,
				},
				user.ErrInvalidCredentials.Error(),
				paths.PathLogin,
			)
		} else {
			log.Error("Password check error during login", map[string]any{"email": email, "error": err.Error()})
			RenderRouteHTML(c, GenericErrorData(c))
		}

		return
	}

	if !foundUser.GetStatus() {
		log.Warn("An inactive user tried to log in", logger.Fields{"mail": email})
		route_utils.RedirectWithError(
			c,
			v,
			map[string]string{
				"email": email,
			},
			user.ErrNotActive.Error(),
			paths.PathLogin,
		)
		return
	}

	session := getSession(c)
	err = foundUser.Login(db, session)

	if err != nil {
		log.Error("Failed to save session after login", map[string]any{"email": email, "error": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	log.Info("Login successful", map[string]any{
		"email":  email,
		"userID": foundUser.GetID(),
	})

	v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: "Successfully logged in!"})
	c.Redirect(http.StatusSeeOther, paths.PathAccount)
}
