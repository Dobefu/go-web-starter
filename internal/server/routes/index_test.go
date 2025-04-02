package routes

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob(filepath.Join("..", "..", "..", "templates", "*.html.tmpl"))
	router.GET("/", Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.NotEmpty(t, w.Body.String())
}
