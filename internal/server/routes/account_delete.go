package routes

import (
	"errors"
	"fmt"
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

func AccountDelete(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	csrfToken := middleware.GetCSRFToken(c)

	formValues := v.GetFormData()

	data := RouteData{
		Template: "pages/account_delete",
		Title:    "Delete Your Account",

		Description: "Confirm your account deletion.",
		HttpStatus:  200,

		FormData: FormData{
			Values: formValues,
			Errors: v.GetSessionErrors(),
		},
		CSRFToken: csrfToken,
	}

	RenderRouteHTML(c, data)
}

func AccountDeletePost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	err := v.ValidateForm(c.Request)

	if err != nil {
		log.Error("Failed to parse form data", map[string]any{"error": err.Error()})
	}

	password := v.GetFormValue(c.Request, "password")

	v.Required("password", password)

	if v.HasErrors() {
		route_utils.RedirectWithError(
			c,
			v,
			nil,
			"Please correct the errors below",
			fmt.Sprintf("%s/delete", paths.PathAccount),
		)
		return
	}

	usr := route_utils.GetUserFromSession(c)

	if usr == nil {
		log.Error("Failed to get user from context", nil)
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	err = usr.CheckPassword(password)

	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			v.AddFieldError("password", user.ErrInvalidCredentials.Error())

			log.Warn("Login failed: invalid credentials (password mismatch)", nil)
			route_utils.RedirectWithError(
				c,
				v,
				nil,
				user.ErrInvalidCredentials.Error(),
				fmt.Sprintf("%s/delete", paths.PathAccount),
			)
		} else {
			log.Error("Password check error during account deletion", nil)
			RenderRouteHTML(c, GenericErrorData(c))
		}

		return
	}

	db, err := route_utils.GetDbFromContext(c)

	if err != nil {
		log.Error("Failed to get database connection from context", nil)
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	_, err = db.Exec("DELETE FROM users WHERE id = $1", usr.GetID())

	if err != nil {
		log.Error("Failed to delete user", logger.Fields{"user_id": usr.GetID})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	session := getSession(c)
	session.Clear()
	_ = session.Save()

	v.SetFlash(message.Message{Type: message.MessageTypeSuccess, Body: "Your account has been deleted successfully."})

	c.Redirect(http.StatusSeeOther, paths.PathLogin)
}
