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

func TestGenericErrorData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.SetFuncMap(server_utils.TemplateFuncMap())
	err := templates.LoadTemplates(router)
	assert.NoError(t, err)

	router.GET("/", func(ctx *gin.Context) {
		RenderRouteHTML(ctx, GenericErrorData(ctx))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
