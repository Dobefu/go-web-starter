package routes

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
	"unsafe"

	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/templates"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"errors"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/user"
	"golang.org/x/crypto/bcrypt"
)

func setupTestRouter(useDBMiddleware bool) (*gin.Engine, sqlmock.Sqlmock, *sql.DB, error) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDB, mockSQL, err := sqlmock.New()

	now := time.Now()
	mockSQL.ExpectQuery(`SELECT (id, username, email, password, status, created_at, updated_at, last_login) FROM users .+`).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"id", "username", "email", "password", "status", "created_at", "updated_at", "last_login"},
			).
				AddRow(1, "username", "test@example.com", "hash", true, now, now, now),
		)

	if err != nil {
		return nil, nil, nil, err
	}

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	if useDBMiddleware {
		router.Use(middleware.Database(mockDB))
	}

	router.SetFuncMap(server_utils.TemplateFuncMap())
	_ = templates.LoadTemplates(router)

	return router, mockSQL, mockDB, nil
}

func TestLoginGET(t *testing.T) {
	router, _, mockDB, err := setupTestRouter(true)
	assert.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	router.GET("/login", Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

type mockUser struct {
	user.User
	checkPasswordFunc func(string) error
}

func (m *mockUser) CheckPassword(password string) error {
	if m.checkPasswordFunc != nil {
		return m.checkPasswordFunc(password)
	}

	return m.User.CheckPassword(password)
}

func (m *mockUser) GetID() int { return 42 }

func setUserPassword(u *user.User, hash string) {
	userVal := reflect.ValueOf(u).Elem()
	passwordField := userVal.FieldByName("password")
	passwordField = reflect.NewAt(passwordField.Type(), unsafe.Pointer(passwordField.UnsafeAddr())).Elem()
	passwordField.SetString(hash)
}

func setupTestRouterWithMocks(t *testing.T, useDBMiddleware bool) (router *gin.Engine, mock sqlmock.Sqlmock, mockDB *sql.DB) {
	gin.SetMode(gin.TestMode)
	router = gin.New()

	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	if useDBMiddleware {
		router.Use(middleware.Database(mockDB))
	}

	router.SetFuncMap(server_utils.TemplateFuncMap())
	_ = templates.LoadTemplates(router)

	return
}

type mockSession struct {
	sessions.Session
	saveErr error
}

func (m *mockSession) Set(key any, val any) {}
func (m *mockSession) Save() error          { return m.saveErr }

func TestLoginPost(t *testing.T) {
	type testCase struct {
		name           string
		setDBInContext bool
		hashPassword   bool
		foundUser      *mockUser
		findUserErr    error
		form           url.Values
		expectStatus   int
		expectLocation string
		checkBody      func(*testing.T, *httptest.ResponseRecorder)
		setupMock      func(router *gin.Engine, mock sqlmock.Sqlmock, tc *testCase)
	}

	origFindByEmail := findByEmail
	origGetSession := getSession

	defer func() {
		findByEmail = origFindByEmail
		getSession = origGetSession
	}()

	tests := []testCase{
		{
			name:           "missing form fields",
			form:           url.Values{},
			expectStatus:   http.StatusSeeOther,
			expectLocation: "/login",
		},
		{
			name:           "FindByEmail returns ErrInvalidCredentials",
			form:           url.Values{"email": {"notfound@example.com"}, "password": {"pw"}},
			setDBInContext: true,
			findUserErr:    user.ErrInvalidCredentials,
			expectStatus:   http.StatusSeeOther,
			expectLocation: "/login",
		},
		{
			name:           "FindByEmail returns DB error",
			form:           url.Values{"email": {"err@example.com"}, "password": {"pw"}},
			setDBInContext: true,
			findUserErr:    errors.New("db fail"),
			expectStatus:   http.StatusInternalServerError,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Server Error")
			},
		},
		{
			name:           "CheckPassword returns ErrInvalidCredentials",
			form:           url.Values{"email": {"user@example.com"}, "password": {"badpw"}},
			hashPassword:   false,
			setDBInContext: true,
			foundUser:      &mockUser{User: *user.NewUser("", "", "", true)},
			expectStatus:   http.StatusSeeOther,
			expectLocation: "/login",
			setupMock: func(router *gin.Engine, mock sqlmock.Sqlmock, tc *testCase) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("notpw"), bcrypt.DefaultCost)
				setUserPassword(&tc.foundUser.User, string(hash))
			},
		},
		{
			name:           "CheckPassword returns other error",
			form:           url.Values{"email": {"user@example.com"}, "password": {"pw"}},
			hashPassword:   false,
			setDBInContext: true,
			foundUser:      &mockUser{User: *user.NewUser("", "", "", true)},
			expectStatus:   http.StatusInternalServerError,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Server Error")
			},
			setupMock: func(router *gin.Engine, mock sqlmock.Sqlmock, tc *testCase) {
				tc.foundUser.checkPasswordFunc = func(string) error {
					return errors.New("bcrypt fail")
				}
			},
		},
		{
			name:           "session save error",
			form:           url.Values{"email": {"user@example.com"}, "password": {"pw"}},
			hashPassword:   true,
			setDBInContext: true,
			foundUser:      &mockUser{User: *user.NewUser("", "", "", true)},
			expectStatus:   http.StatusInternalServerError,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Server Error")
			},
			setupMock: func(router *gin.Engine, mock sqlmock.Sqlmock, tc *testCase) {
				getSession = func(c *gin.Context) sessions.Session {
					return &mockSession{saveErr: errors.New("session fail")}
				}
			},
		},
		{
			name:           "success",
			form:           url.Values{"email": {"user@example.com"}, "password": {"pw"}},
			hashPassword:   true,
			setDBInContext: true,
			foundUser:      &mockUser{User: *user.NewUser("", "", "", true)},
			expectStatus:   http.StatusSeeOther,
			expectLocation: "/",
			setupMock: func(router *gin.Engine, mock sqlmock.Sqlmock, tc *testCase) {
				userVal := reflect.ValueOf(&tc.foundUser.User).Elem()
				idField := userVal.FieldByName("id")
				reflect.NewAt(idField.Type(), unsafe.Pointer(idField.UnsafeAddr())).Elem().SetInt(42)

				mock.ExpectQuery(`UPDATE users SET username = \$1, email = \$2, password = \$3, status = \$4, updated_at = \$5, last_login = \$6 WHERE id = \$7 RETURNING updated_at`).
					WithArgs("", "", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg(), 42).
					WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(time.Now()))
			},
		},
		{
			name:           "inactive",
			form:           url.Values{"email": {"user@example.com"}, "password": {"pw"}},
			hashPassword:   true,
			setDBInContext: true,
			foundUser:      &mockUser{},
			expectStatus:   http.StatusSeeOther,
			expectLocation: "/login",
		},
		{
			name:           "ValidateForm error",
			form:           url.Values{"email": {"user@example.com"}, "password": {"pw"}},
			hashPassword:   false,
			setDBInContext: true,
			foundUser:      &mockUser{User: *user.NewUser("", "", "", true)},
			findUserErr:    nil,
			expectStatus:   http.StatusSeeOther,
			expectLocation: "/login",
		},
		{
			name:           "db in context but wrong type",
			form:           url.Values{"email": {"user@example.com"}, "password": {"pw"}},
			hashPassword:   false,
			setDBInContext: false,
			expectStatus:   http.StatusInternalServerError,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Server Error")
			},
			setupMock: func(router *gin.Engine, mock sqlmock.Sqlmock, tc *testCase) {
				router.Use(func(c *gin.Context) {
					c.Set("db", 123)
					c.Next()
				})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router, mock, mockDB := setupTestRouterWithMocks(t, tc.setDBInContext)
			defer func() { _ = mockDB.Close() }()

			findByEmail = func(db database.DatabaseInterface, email string) (*user.User, error) {
				if tc.foundUser != nil {
					return &tc.foundUser.User, tc.findUserErr
				}

				return nil, tc.findUserErr
			}

			getSession = origGetSession

			w := httptest.NewRecorder()
			var req *http.Request

			if tc.name == "ValidateForm error" {
				req, _ = http.NewRequest("POST", "/login", strings.NewReader("%%%"))
			} else {
				req, _ = http.NewRequest("POST", "/login", strings.NewReader(tc.form.Encode()))
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			if tc.hashPassword && tc.foundUser != nil {
				hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
				setUserPassword(&tc.foundUser.User, string(hash))
			}

			if tc.setupMock != nil {
				tc.setupMock(router, mock, &tc)
			}

			router.POST("/login", LoginPost)

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectStatus, w.Code)

			if tc.expectLocation != "" {
				assert.Equal(t, tc.expectLocation, w.Header().Get("Location"))
			}

			if tc.checkBody != nil {
				tc.checkBody(t, w)
			}
		})
	}
}
