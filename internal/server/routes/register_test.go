package routes

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Dobefu/go-web-starter/internal/database"
	email "github.com/Dobefu/go-web-starter/internal/email"
	"github.com/Dobefu/go-web-starter/internal/server/routes/paths"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	netSmtp "net/smtp"
)

func setupRouter(db database.DatabaseInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	store := cookie.NewStore([]byte("secret"))

	router.Use(sessions.Sessions("mysession", store))

	router.POST(paths.PathRegister, func(c *gin.Context) {
		c.Set("db", db)
		RegisterPost(c)
	})

	return router
}

func patchFinders(usernameFn, emailFn func(database.DatabaseInterface, string) (*user.User, error)) (restore func()) {
	origFindByUsername := findByUsername
	origFindByEmail := findByEmail

	findByUsername = usernameFn
	findByEmail = emailFn

	return func() {
		findByUsername = origFindByUsername
		findByEmail = origFindByEmail
	}
}

func patchEmailer() (restore func()) {
	origSendMail := email.SmtpSendMail

	email.SmtpSendMail = func(addr string, a netSmtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}

	return func() { email.SmtpSendMail = origSendMail }
}

func makeForm(fields map[string]string) url.Values {
	v := url.Values{}

	for k, val := range fields {
		v.Set(k, val)
	}

	return v
}

func makeRequest(router *gin.Engine, form url.Values) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", paths.PathRegister, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	router.ServeHTTP(w, req)

	return w
}

func TestRegisterPost(t *testing.T) {
	now := time.Now()

	viper.Set("site.name", "Test Site")
	viper.Set("site.host", "http://localhost:8080")

	defaultFind := func(db database.DatabaseInterface, s string) (*user.User, error) { return nil, errors.New("not found") }
	userTaken := func(db database.DatabaseInterface, s string) (*user.User, error) { return &user.User{}, nil }

	tests := []struct {
		name           string
		fields         map[string]string
		findByUsername func(database.DatabaseInterface, string) (*user.User, error)
		findByEmail    func(database.DatabaseInterface, string) (*user.User, error)
		mockDB         bool
		mockSuccessDB  bool
		expectStatus   int
		expectLocation string
	}{
		{
			name:           "missing form fields",
			fields:         map[string]string{},
			findByUsername: defaultFind,
			findByEmail:    defaultFind,
			mockDB:         true,
			expectStatus:   http.StatusSeeOther,
		},
		{
			name:           "username taken",
			fields:         map[string]string{"username": "taken", "email": "test@example.com", "password": "password123", "password_confirm": "password123"},
			findByUsername: userTaken,
			findByEmail:    defaultFind,
			mockDB:         true,
			expectStatus:   http.StatusSeeOther,
		},
		{
			name:           "email taken",
			fields:         map[string]string{"username": "user", "email": "taken@example.com", "password": "password123", "password_confirm": "password123"},
			findByUsername: defaultFind,
			findByEmail:    userTaken,
			mockDB:         true,
			expectStatus:   http.StatusSeeOther,
		},
		{
			name:           "passwords do not match",
			fields:         map[string]string{"username": "user", "email": "test@example.com", "password": "password123", "password_confirm": "password321"},
			findByUsername: defaultFind,
			findByEmail:    defaultFind,
			mockDB:         true,
			expectStatus:   http.StatusSeeOther,
		},
		{
			name:           "success",
			fields:         map[string]string{"username": "user", "email": "test@example.com", "password": "password123", "password_confirm": "password123"},
			findByUsername: defaultFind,
			findByEmail:    defaultFind,
			mockSuccessDB:  true,
			expectStatus:   http.StatusSeeOther,
			expectLocation: fmt.Sprintf("%s/verify?email=test@example.com", paths.PathRegister),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			restoreFinders := patchFinders(tc.findByUsername, tc.findByEmail)
			defer restoreFinders()

			var db database.DatabaseInterface
			var mock sqlmock.Sqlmock

			if tc.mockSuccessDB {
				dbSQL, m, err := sqlmock.New()
				assert.NoError(t, err)
				defer func() { _ = dbSQL.Close() }()

				m.ExpectQuery("INSERT INTO users").
					WithArgs("user", "test@example.com", sqlmock.AnyArg(), false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "last_login"}).AddRow(1, now, now, now))

				db = dbSQL

				restoreEmail := patchEmailer()
				defer restoreEmail()

				mock = m
			} else if tc.mockDB {
				db = &struct{ database.DatabaseInterface }{}
			}

			router := setupRouter(db)
			w := makeRequest(router, makeForm(tc.fields))

			if tc.expectLocation != "" {
				assert.Contains(t, w.Header().Get("Location"), tc.expectLocation)
			}

			assert.Equal(t, tc.expectStatus, w.Code)

			if tc.mockSuccessDB {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
