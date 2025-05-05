package routes

import (
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

	if _, ok := formValues["email"]; !ok {
		formValues["email"] = currentUser.GetEmail()
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
