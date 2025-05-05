package utils

import (
	"net/http"

	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

func RedirectWithError(
	c *gin.Context,
	v *validator.Validator,
	fields map[string]string,
	flashMsg string,
	path string,
) {
	v.SetFormData(fields)

	v.SetErrors()
	v.SetFlash(message.Message{Body: flashMsg})
	c.Redirect(http.StatusSeeOther, path)
}
