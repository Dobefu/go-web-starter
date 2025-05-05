package routes

import (
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

func AccountEdit(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	csrfToken := middleware.GetCSRFToken(c)

	currentUser := route_utils.GetUserFromSession(c)

	formValues := v.GetFormData()

	if _, ok := formValues["username"]; !ok {
		formValues["username"] = currentUser.GetUsername()
	}

	data := RouteData{
		Template: "pages/account_edit",
		Title:    "Edit Your Account",

		Description: "Manage your account details.",
		HttpStatus:  200,

		FormData: FormData{
			Values: formValues,
			Errors: v.GetSessionErrors(),
		},
		CSRFToken: csrfToken,
	}

	RenderRouteHTML(c, data)
}

func AccountEditPost(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	err := v.ValidateForm(c.Request)

	if err != nil {
		log.Error("Failed to parse form data", map[string]any{"error": err.Error()})
	}

	username := v.GetFormValue(c.Request, "username")

	v.Required("username", username)
	v.MinLength("username", username, 3)

	usr := route_utils.GetUserFromSession(c)
	db, err := route_utils.GetDbFromContext(c)

	if err != nil {
		log.Error("Could not get the database from the context", logger.Fields{"error": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	_, err = findByUsername(db, username)

	if err == nil {
		v.AddFieldError("username", "This username is already taken")
	}

	if v.HasErrors() {
		redirectToAccountEditWithError(c, v, username, "Please correct the errors below")
		return
	}

	usr.SetUsername(username)
	err = usr.Save(db)

	if err != nil {
		log.Error("Could not update the user", logger.Fields{"error": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	v.SetFlash(message.Message{
		Type: message.MessageTypeSuccess,
		Body: "Your profile has been updated successfully!",
	})

	c.Redirect(http.StatusSeeOther, "/account")
}

func redirectToAccountEditWithError(c *gin.Context, v *validator.Validator, username string, flashMsg string) {
	v.SetFormData(map[string]string{
		"username": username,
	})

	v.SetErrors()
	v.SetFlash(message.Message{Body: flashMsg})
	c.Redirect(http.StatusSeeOther, "/account/edit")
}
