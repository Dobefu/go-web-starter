package routes

import (
	"net/http"

	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

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
