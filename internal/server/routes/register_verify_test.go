package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRegisterVerify(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "without query params",
			path:       "/",
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "with email only",
			path:       "/?email=test@example.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "with email and token",
			path:       "/?email=test@example.com&token=test-token",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("site.name", "Test Site")
			viper.Set("site.host", "http://localhost:8080")

			router := gin.New()
			store := cookie.NewStore([]byte("secret"))
			router.Use(sessions.Sessions("mysession", store))
			router.SetFuncMap(server_utils.TemplateFuncMap())
			err := templates.LoadTemplates(router)
			assert.NoError(t, err)
			router.GET("/", RegisterVerify)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
