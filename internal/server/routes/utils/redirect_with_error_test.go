package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestRedirectWithError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	store := cookie.NewStore([]byte("secret"))
	sessions.Sessions("mysession", store)(c)

	v := validator.New()
	v.SetContext(c)

	router := gin.New()
	router.GET("/", func(ctx *gin.Context) {
		RedirectWithError(c, v, map[string]string{}, "test-message", "/path")
	})

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusSeeOther, w.Code)
}
