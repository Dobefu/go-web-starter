package routes

import (
	"errors"
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var findByEmail = user.FindByEmail
var getSession = sessions.Default

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

	v.Required("email", email)
	v.Required("password", password)

	if v.HasErrors() {
		redirectToLoginWithError(c, v, email, "Please correct the errors below")
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
			log.Warn("Login failed: invalid credentials (email not found)", map[string]any{"email": email})
			redirectToLoginWithError(c, v, email, user.ErrInvalidCredentials.Error())
		} else {
			log.Error("Database error during login", map[string]any{"email": email, "error": err.Error()})
			RenderRouteHTML(c, GenericErrorData(c))
		}

		return
	}

	err = foundUser.CheckPassword(password)

	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			log.Warn("Login failed: invalid credentials (password mismatch)", map[string]any{"email": email})
			redirectToLoginWithError(c, v, email, user.ErrInvalidCredentials.Error())
		} else {
			log.Error("Password check error during login", map[string]any{"email": email, "error": err.Error()})
			RenderRouteHTML(c, GenericErrorData(c))
		}

		return
	}

	session := getSession(c)
	session.Set("userID", foundUser.GetID())

	err = session.Save()

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
	c.Redirect(http.StatusSeeOther, "/")
}

func redirectToLoginWithError(c *gin.Context, v *validator.Validator, email string, flashMsg string) {
	v.SetFormData(map[string]string{
		"email": email,
	})

	if !v.HasErrors() {
		v.AddFieldError("email", user.ErrInvalidCredentials.Error())
		v.AddFieldError("password", " ")
	}

	v.SetErrors()
	v.SetFlash(message.Message{Body: flashMsg})
	c.Redirect(http.StatusSeeOther, "/login")
}
