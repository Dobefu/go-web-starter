package routes

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
)

func setupTestRouter() (*gin.Engine, sqlmock.Sqlmock, *sql.DB, error) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDB, mockSQL, err := sqlmock.New()

	if err != nil {
		return nil, nil, nil, err
	}

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.Use(middleware.Database(mockDB))
	router.SetFuncMap(server_utils.TemplateFuncMap())
	_ = templates.LoadTemplates(router)

	return router, mockSQL, mockDB, nil
}

func TestLoginGET(t *testing.T) {
	router, _, mockDB, err := setupTestRouter()
	assert.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	router.GET("/", Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginPostSuccess(t *testing.T) {
	router, mockSQL, mockDB, err := setupTestRouter()
	assert.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	router.POST("/login", LoginPost)

	expectedQuery := "SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE email = \\$1"

	mockSQL.ExpectQuery(expectedQuery).
		WithArgs("test@example.com").
		WillReturnError(sql.ErrNoRows)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "validpassword123")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	err = mockSQL.ExpectationsWereMet()
	assert.NoError(t, err, "SQL expectations were not met: %v", err)
}

func TestLoginPostParse(t *testing.T) {
	router, _, mockDB, err := setupTestRouter()
	assert.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	router.POST("/login", LoginPost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}
