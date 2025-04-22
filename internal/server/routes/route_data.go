package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FormData struct {
	Values   map[string]string
	Errors   map[string][]string
	Messages []string
}

type RouteData struct {
	Template    string
	HttpStatus  int
	Title       string
	Description string
	Data        map[string]any
	FormData    FormData
	CSRFToken   string
}

func GenericErrorData(c *gin.Context) RouteData {
	return RouteData{
		Template:    "pages/500",
		HttpStatus:  http.StatusInternalServerError,
		Title:       "Server Error",
		Description: "Sorry, something went wrong on our end.",
	}
}
