package routes

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRenderRouteHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)

	viper.Set("site.name", "Test Site")
	viper.Set("site.host", "http://localhost:8080")

	t.Run("Debug Mode Rendering", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)
		defer gin.SetMode(gin.TestMode)

		routeData := RouteData{
			Template:   "pages/index",
			HttpStatus: http.StatusOK,
			Title:      "Test Page",
		}

		err := templates.LoadTemplates(r)
		assert.NoError(t, err)

		RenderRouteHTML(c, routeData)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Page")
	})

	t.Run("Cached Template Rendering", func(t *testing.T) {
		tmpl := template.Must(template.New("test.gohtml").Parse("Hello {{.Title}}"))
		cache := templates.GetTemplateCache()
		cache.Set("test.gohtml", tmpl)

		routeData := RouteData{
			Template:   "test.gohtml",
			HttpStatus: http.StatusOK,
			Title:      "Test Page",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)
		sessions.Sessions("mysession", store)(c)

		RenderRouteHTML(c, routeData)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello Test Page")
	})

	t.Run("Template Execution Error", func(t *testing.T) {
		tmpl := template.Must(template.New("error.gohtml").Parse("{{.BogusField}}"))
		cache := templates.GetTemplateCache()
		cache.Set("error.gohtml", tmpl)

		routeData := RouteData{
			Template:   "error.gohtml",
			HttpStatus: http.StatusOK,
			Title:      "Test Page",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)
		sessions.Sessions("mysession", store)(c)

		RenderRouteHTML(c, routeData)

		assert.NotNil(t, c.Errors.Last())
	})
}
