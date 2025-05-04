package routes

import (
	"net/http"

	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

func RegisterVerify(c *gin.Context) {
	v := validator.New()
	v.SetContext(c)

	data := RouteData{
		Template:   "pages/register_verify",
		HttpStatus: http.StatusOK,

		Title:       "Verify email address",
		Description: "Verify your email address.",

		Data: map[string]any{
			"Token": "",
			"Email": "your email address",
		},
	}

	data.Data["Token"] = v.GetFormValue(c.Request, "token")
	email := v.GetFormValue(c.Request, "email")

	if len(email) > 0 {
		data.Data["Email"] = email
	}

	RenderRouteHTML(c, data)
}
