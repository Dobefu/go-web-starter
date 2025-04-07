package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.SetFuncMap(server_utils.TemplateFuncMap())
	err := templates.LoadTemplates(router)
	assert.NoError(t, err)
	router.GET("/bogus", NotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bogus", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
