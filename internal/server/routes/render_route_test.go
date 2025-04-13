package routes

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRenderRouteHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/", nil)
	c.Request = req

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
		tmpl := template.Must(template.New("test.tmpl").Parse("Hello {{.Title}}"))
		cache := templates.GetTemplateCache()
		cache.Set("test.tmpl", tmpl)

		routeData := RouteData{
			Template:   "test.tmpl",
			HttpStatus: http.StatusOK,
			Title:      "Test Page",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("GET", "/", nil)
		c.Request = req
		RenderRouteHTML(c, routeData)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello Test Page")
	})

	t.Run("Template Execution Error", func(t *testing.T) {
		tmpl := template.Must(template.New("error.tmpl").Parse("{{.BogusField}}"))
		cache := templates.GetTemplateCache()
		cache.Set("error.tmpl", tmpl)

		routeData := RouteData{
			Template:   "error.tmpl",
			HttpStatus: http.StatusOK,
			Title:      "Test Page",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("GET", "/", nil)
		c.Request = req
		RenderRouteHTML(c, routeData)

		assert.NotNil(t, c.Errors.Last())
	})
}
