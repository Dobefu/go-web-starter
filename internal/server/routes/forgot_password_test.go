package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	viper.Set("site.name", "Test Site")
	viper.Set("site.host", "http://localhost:8080")

	router := gin.New()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	router.SetFuncMap(server_utils.TemplateFuncMap())
	err := templates.LoadTemplates(router)
	assert.NoError(t, err)
	router.GET(paths.PathForgotPassword, ForgotPassword)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", paths.PathForgotPassword, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
