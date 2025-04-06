package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMinifyText(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Minify())

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte(" some     text"))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)

	assert.Equal(t, " some     text", string(body))
}

func TestMinifyJson(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Minify())

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(` { "testing": true } `))
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)

	assert.Equal(t, `{"testing":true}`, string(body))
}

func TestMinifyHtml(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Minify())

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
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)

	assert.Equal(
		t,
		"<!doctype html><html lang=en><head><meta charset=UTF-8></head><body><p>Some paragraph</body></html>",
		string(body),
	)
}
