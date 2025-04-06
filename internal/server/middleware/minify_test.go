package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	*httptest.ResponseRecorder
	size    int
	status  int
	written bool
}

func (m *mockWriter) CloseNotify() <-chan bool {
	return make(chan bool, 1)
}

func (m *mockWriter) Pusher() http.Pusher {
	return nil
}

func (m *mockWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hijack not supported")
}

func (m *mockWriter) Size() int {
	return m.size
}

func (m *mockWriter) Status() int {
	return m.status
}

func (m *mockWriter) Written() bool {
	return m.written
}

func (m *mockWriter) WriteHeaderNow() {
	if !m.written {
		m.WriteHeader(m.status)
		m.written = true
	}
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Minify())

	return router
}

func TestMinifyText(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte(" some     text"))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, " some     text", string(body))
}

func TestMinifyJson(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(` { "testing": true } `))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"testing":true}`, string(body))
}

func TestMinifyHtml(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html", []byte(`
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
  </head>
  <body>
    <p>  Some paragraph</p>
  </body>
</html>
		`))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, "<!doctype html><html lang=en><head><meta charset=UTF-8></head><body><p>Some paragraph</body></html>", string(body))
}

func TestMinifyInvalidJson(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	invalidJSON := `{"invalid": true, "missing": value}`
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(invalidJSON))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, invalidJSON, string(body))
}

func TestMinifyUnsupportedContentType(t *testing.T) {
	t.Parallel()
	router := newTestRouter()

	content := "some content"
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/custom", []byte(content))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, content, string(body))
}

func TestResponseWriterWrite(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	originalWriter := &mockWriter{
		ResponseRecorder: httptest.NewRecorder(),
		size:             0,
		status:           http.StatusOK,
		written:          false,
	}
	rw := &ResponseWriter{
		ResponseWriter: originalWriter,
		body:           buf,
	}

	testData := []byte("test data")
	n, err := rw.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, buf.Bytes())
}
