package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.SetFuncMap(server_utils.TemplateFuncMap())
	templates.LoadTemplates(router)

	return router
}

func TestLoginGET(t *testing.T) {
	router := setupTestRouter()
	router.GET("/", Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginPostSuccess(t *testing.T) {
	router := setupTestRouter()
	router.POST("/login", LoginPost)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "validpassword123")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

func TestLoginPostParse(t *testing.T) {
	router := setupTestRouter()
	router.POST("/login", LoginPost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}
