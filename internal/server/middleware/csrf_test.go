package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCSRFTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("test-session", store))

	return r
}

func setupTestRequest(method, path string, formValues url.Values) (*http.Request, error) {
	var body string

	if formValues != nil {
		body = formValues.Encode()
	}

	req, err := http.NewRequest(method, path, strings.NewReader(body))

	if err != nil {
		return nil, err
	}

	if formValues != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}

func addSessionCookies(req *http.Request, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

func TestCSRFMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		csrfToken      string
		expectedStatus int
	}{
		{
			name:           "GET request should pass through",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with valid token should pass",
			method:         "POST",
			csrfToken:      "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with invalid token should redirect",
			method:         "POST",
			csrfToken:      "invalid-token",
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "POST with missing token should redirect",
			method:         "POST",
			csrfToken:      "",
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupCSRFTestRouter()

			w := httptest.NewRecorder()
			req, err := setupTestRequest("GET", "/setup", nil)
			assert.NoError(t, err)

			router.GET("/setup", func(c *gin.Context) {
				if tt.csrfToken != "" {
					session := sessions.Default(c)
					session.Set("csrf_token", "valid-token")
					session.Save()
				}

				c.Status(http.StatusOK)
			})

			router.ServeHTTP(w, req)

			cookies := w.Result().Cookies()

			router.Use(CSRF())

			w = httptest.NewRecorder()
			formValues := url.Values{}

			if tt.method == "POST" {
				formValues.Add("_csrf", tt.csrfToken)
			}

			req, err = setupTestRequest(tt.method, "/test", formValues)
			assert.NoError(t, err)
			addSessionCookies(req, cookies)

			router.Any("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetCSRFToken(t *testing.T) {
	t.Run("generates and caches token", func(t *testing.T) {
		router := setupCSRFTestRouter()

		var token1, token2 string

		router.GET("/test", func(c *gin.Context) {
			token1 = GetCSRFToken(c)
			assert.NotEmpty(t, token1)
			assert.Len(t, token1, 44)

			token2 = GetCSRFToken(c)
			assert.Equal(t, token1, token2)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, err := setupTestRequest("GET", "/test", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, token1)
		assert.Equal(t, token1, token2)
	})

	t.Run("returns empty string on rand.Read error", func(t *testing.T) {
		originalRandRead := csrfRandRead
		defer func() { csrfRandRead = originalRandRead }()
		csrfRandRead = func(b []byte) (n int, err error) {
			return 0, assert.AnError
		}

		router := setupCSRFTestRouter()
		var token string

		router.GET("/test", func(c *gin.Context) {
			token = GetCSRFToken(c)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, err := setupTestRequest("GET", "/test", nil)
		assert.NoError(t, err)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, token)
	})
}
