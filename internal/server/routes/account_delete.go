package routes

import (
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
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
