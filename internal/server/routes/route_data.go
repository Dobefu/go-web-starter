package routes

import (
	"net/http"

	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/gin-gonic/gin"
)

type FormData struct {
	Values map[string]string
	Errors map[string][]string
}

type RouteData struct {
	Template    string
	HttpStatus  int
	Title       string
	Description string
	Data        map[string]any
	FormData    FormData
	CSRFToken   string
	User        *user.User
}

func GenericErrorData(c *gin.Context) RouteData {
	return RouteData{
		Template:    "pages/server-error",
		HttpStatus:  http.StatusInternalServerError,
		Title:       "Server Error",
		Description: "Sorry, something went wrong on our end.",
	}
}
