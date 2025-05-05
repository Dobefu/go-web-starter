package utils

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

type mockDB struct{}

func (d *mockDB) Close() error                                       { return nil }
func (d *mockDB) Ping() error                                        { return nil }
func (d *mockDB) Query(query string, args ...any) (*sql.Rows, error) { return nil, nil }
func (d *mockDB) QueryRow(query string, args ...any) *sql.Row        { return nil }
func (d *mockDB) Exec(query string, args ...any) (sql.Result, error) { return nil, nil }
func (d *mockDB) Begin() (*sql.Tx, error)                            { return nil, nil }
func (d *mockDB) Stats() sql.DBStats                                 { return sql.DBStats{} }

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
			assert.Equal(t, tc.expect, GetUserFromSession(c))
		})
	}
}
