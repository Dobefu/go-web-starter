package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		handler        gin.HandlerFunc
	}{
		{
			name:           "successful request",
			path:           "/",
			method:         "GET",
			expectedStatus: http.StatusOK,
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
		},
		{
			name:           "query parameters",
			path:           "/?q=test",
			method:         "GET",
			expectedStatus: http.StatusOK,
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router.GET(tt.path, tt.handler)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, tt.path, nil)
			assert.NoError(t, err)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
