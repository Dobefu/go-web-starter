package routes

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"database/sql"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/user"
)

type mockDB struct{}

func (d *mockDB) Close() error                                       { return nil }
func (d *mockDB) Ping() error                                        { return nil }
func (d *mockDB) Query(query string, args ...any) (*sql.Rows, error) { return nil, nil }
func (d *mockDB) QueryRow(query string, args ...any) *sql.Row        { return nil }
func (d *mockDB) Exec(query string, args ...any) (sql.Result, error) { return nil, nil }
func (d *mockDB) Begin() (*sql.Tx, error)                            { return nil, nil }
func (d *mockDB) Stats() sql.DBStats                                 { return sql.DBStats{} }

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

func TestGetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	origUserFindByID := userFindByID
	defer func() { userFindByID = origUserFindByID }()

	mockedUser := &user.User{}

	testCases := []struct {
		name        string
		userID      any
		db          any
		patchFindBy func()
		expect      *user.User
	}{
		{
			"no user in session",
			nil,
			nil,
			func() { userFindByID = origUserFindByID },
			nil,
		},
		{
			"no db in context",
			1,
			nil,
			func() { userFindByID = origUserFindByID },
			nil,
		},
		{
			"db wrong type",
			1,
			struct{}{},
			func() { userFindByID = origUserFindByID },
			nil,
		},
		{
			"user not found in db",
			1,
			&mockDB{},
			func() {
				userFindByID = func(db database.DatabaseInterface, id int) (*user.User, error) {
					return nil, errors.New("not found")
				}
			},
			nil,
		},
		{
			"user found",
			1,
			&mockDB{},
			func() {
				userFindByID = func(db database.DatabaseInterface, id int) (*user.User, error) {
					return mockedUser, nil
				}
			},
			mockedUser,
		},
		{
			"userID cannot be parsed to int",
			"not-an-int",
			&mockDB{},
			func() { userFindByID = origUserFindByID },
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request, _ = http.NewRequest("GET", "/", nil)
			sessions.Sessions("mysession", cookie.NewStore([]byte("secret")))(c)

			if tc.userID != nil {
				sess := sessions.Default(c)
				sess.Set("userID", tc.userID)

				_ = sess.Save()
			}

			if tc.db != nil {
				c.Set("db", tc.db)
			}

			tc.patchFindBy()
			assert.Equal(t, tc.expect, getCurrentUser(c))
		})
	}
}
