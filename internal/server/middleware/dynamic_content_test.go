package middleware

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockResponseWriter struct {
	*httptest.ResponseRecorder
}

func (m *mockResponseWriter) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func (m *mockResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (m *mockResponseWriter) Pusher() http.Pusher {
	return nil
}

func (m *mockResponseWriter) Size() int {
	return m.ResponseRecorder.Body.Len()
}

func (m *mockResponseWriter) Status() int {
	return m.ResponseRecorder.Code
}

func (m *mockResponseWriter) WriteHeaderNow() {
	// Placeholder function for testing.
}

func (m *mockResponseWriter) Written() bool {
	return m.ResponseRecorder.Code != 0
}

func TestDynamicContentDebugMode(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	defer gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(DynamicContent())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "test content")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test content", w.Body.String())
}

func TestDynamicContentExistingResponseWriter(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(DynamicContent())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "test content")
	})

	recorder := httptest.NewRecorder()

	w := &ResponseWriter{
		ResponseWriter: &mockResponseWriter{recorder},
		body:           new(bytes.Buffer),
	}

	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "test content", w.body.String())
}

func TestDynamicContentNewResponseWriter(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(DynamicContent())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "test content")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test content", w.Body.String())
}

func TestDynamicContentDuplicateMiddleware(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(DynamicContent())
	router.Use(DynamicContent())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "test content")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
