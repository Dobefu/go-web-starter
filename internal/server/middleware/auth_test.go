package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouterWithSession(mw gin.HandlerFunc, route string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))

	r.Use(sessions.Sessions("test-session", store))
	r.Use(mw)
	r.GET(route, handler)

	return r
}

func getAuthCookies() []*http.Cookie {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))

	r.Use(sessions.Sessions("test-session", store))

	r.GET("/set", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("userID", "123")
		_ = session.Save()
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/set", nil)
	r.ServeHTTP(w, req)

	return w.Result().Cookies()
}

func TestAuthOnly(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		expectedCode  int
		expectedLoc   string
	}{
		{"redirects to login if not authenticated", false, http.StatusSeeOther, paths.PathLogin},
		{"allows access if authenticated", true, http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouterWithSession(AuthOnly(), "/protected", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)

			if tt.authenticated {
				for _, cookie := range getAuthCookies() {
					req.AddCookie(cookie)
				}
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedLoc != "" {
				assert.Equal(t, tt.expectedLoc, w.Header().Get("Location"))
			}
		})
	}
}

func TestAnonOnly(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		expectedCode  int
		expectedLoc   string
	}{
		{"redirects to home if authenticated", true, http.StatusSeeOther, paths.PathAccount},
		{"allows access if not authenticated", false, http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouterWithSession(AnonOnly(), paths.PathLogin, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", paths.PathLogin, nil)

			if tt.authenticated {
				for _, cookie := range getAuthCookies() {
					req.AddCookie(cookie)
				}
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedLoc != "" {
				assert.Equal(t, tt.expectedLoc, w.Header().Get("Location"))
			}
		})
	}
}
