package user

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	testUserID        = 1
	testUsername      = "Test User"
	testEmail         = "test@user.com"
	testCreatedAtUnix = 100000
	testUpdatedAtUnix = 200000
)

type mockDatabase struct {
	mock sqlmock.Sqlmock
	db   *sql.DB
}

func (m *mockDatabase) Close() error {
	return nil
}

func (m *mockDatabase) Ping() error {
	return nil
}

func (m *mockDatabase) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockDatabase) QueryRow(query string, args ...any) *sql.Row {
	driverArgs := make([]driver.Value, len(args))

	for i, arg := range args {
		driverArgs[i] = arg
	}

	return m.db.QueryRow(query, args...)
}

func (m *mockDatabase) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *mockDatabase) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockDatabase) Stats() sql.DBStats {
	return sql.DBStats{}
}

func setupUserTests() (user User) {
	return User{
		id:        testUserID,
		username:  testUsername,
		email:     testEmail,
		status:    true,
		createdAt: time.Unix(testCreatedAtUnix, 0),
		updatedAt: time.Unix(testUpdatedAtUnix, 0),
	}
}

func TestUser_Getters_TableDriven(t *testing.T) {
	user := setupUserTests()
	tests := []struct {
		name   string
		getter func() any
		want   any
	}{
		{"GetID", func() any { return user.GetID() }, testUserID},
		{"GetUsername", func() any { return user.GetUsername() }, testUsername},
		{"GetEmail", func() any { return user.GetEmail() }, testEmail},
		{"GetStatus", func() any { return user.GetStatus() }, true},
		{"GetCreatedAt", func() any { return user.GetCreatedAt() }, time.Unix(testCreatedAtUnix, 0)},
		{"GetUpdatedAt", func() any { return user.GetUpdatedAt() }, time.Unix(testUpdatedAtUnix, 0)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.getter())
		})
	}
}

func TestUser_Save_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock, user *User, now time.Time)
		wantErr   string
	}{
		{
			name: "query row error",
			mockSetup: func(mock sqlmock.Sqlmock, user *User, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
					WithArgs(user.username, user.email, user.password, user.status, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone.Error(),
		},
		{
			name: "scan error",
			mockSetup: func(mock sqlmock.Sqlmock, user *User, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
					WithArgs(user.username, user.email, user.password, user.status, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(nil, nil, nil).RowError(0, sql.ErrNoRows))
			},
			wantErr: sql.ErrNoRows.Error(),
		},
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock, user *User, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
					WithArgs(user.username, user.email, user.password, user.status, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(99, now, now))
			},
			wantErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			now := time.Now()
			user := setupUserTests()
			tc.mockSetup(mock, &user, now)
			err := user.Save(db)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 99, user.GetID())
				assert.WithinDuration(t, now, user.GetCreatedAt(), time.Second)
				assert.WithinDuration(t, now, user.GetUpdatedAt(), time.Second)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCheckPassword_TableDriven(t *testing.T) {
	hashed, err := HashPassword("supersecret")
	assert.NoError(t, err)

	tests := []struct {
		name         string
		userPassword string
		input        string
		wantErr      error
		wantOtherErr bool
	}{
		{
			name:         "success",
			userPassword: hashed,
			input:        "supersecret",
			wantErr:      nil,
		},
		{
			name:         "invalid password",
			userPassword: hashed,
			input:        "wrongpassword",
			wantErr:      ErrInvalidCredentials,
		},
		{
			name:         "not a hash",
			userPassword: "notahash",
			input:        "irrelevant",
			wantErr:      nil,
			wantOtherErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user := setupUserTests()
			user.password = tc.userPassword
			err := user.CheckPassword(tc.input)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else if tc.wantOtherErr {
				assert.Error(t, err)
				assert.NotErrorIs(t, err, ErrInvalidCredentials)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{"success", "testpass", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hashed, err := HashPassword(tc.input)

			if tc.wantOK {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashed)
				assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashed), []byte(tc.input)))
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewUser_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		hash     string
		status   bool
	}{
		{"basic", "newuser", "new@user.com", "$2a$10$somethinghashed", true},
		{"inactive", "inactive", "inactive@user.com", "$2a$10$inactivehash", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user := NewUser(tc.username, tc.email, tc.hash, tc.status)
			assert.Equal(t, tc.username, user.username)
			assert.Equal(t, tc.email, user.email)
			assert.Equal(t, tc.hash, user.password)
			assert.Equal(t, tc.status, user.status)
		})
	}
}

func setupMockDB(t *testing.T) (*mockDatabase, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	cleanup := func() { _ = db.Close() }

	return &mockDatabase{db: db, mock: mock}, mock, cleanup
}

func userRow(id int, username, email, password string, status bool, createdAt, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.
		NewRows([]string{"id", "username", "email", "password", "status", "created_at", "updated_at"}).
		AddRow(id, username, email, password, status, createdAt, updatedAt)
}

func TestFindByEmail_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock, now time.Time)
		expectErr error
		expectNil bool
	}{
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs(testEmail).
					WillReturnRows(userRow(testUserID, testUsername, testEmail, "hash", true, now, now))
			},
			expectErr: nil,
			expectNil: false,
		},
		{
			name: "not found",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs(testEmail).
					WillReturnError(sql.ErrNoRows)
			},
			expectErr: ErrInvalidCredentials,
			expectNil: true,
		},
		{
			name: "db error",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs(testEmail).
					WillReturnError(sql.ErrConnDone)
			},
			expectErr: sql.ErrConnDone,
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			now := time.Now()
			tc.mockSetup(mock, now)
			user, err := FindByEmail(db, testEmail)

			if tc.expectNil {
				assert.Nil(t, user)
			} else {
				assert.NotNil(t, user)
			}

			if tc.expectErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFindByID_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock, now time.Time)
		expectErr string
		expectNil bool
	}{
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE id = $1`)).
					WithArgs(testUserID).
					WillReturnRows(userRow(testUserID, testUsername, testEmail, "hash", true, now, now))
			},
			expectErr: "",
			expectNil: false,
		},
		{
			name: "not found",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE id = $1`)).
					WithArgs(testUserID).
					WillReturnError(sql.ErrNoRows)
			},
			expectErr: "not found",
			expectNil: true,
		},
		{
			name: "db error",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE id = $1`)).
					WithArgs(testUserID).
					WillReturnError(sql.ErrConnDone)
			},
			expectErr: "error finding user by ID",
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			now := time.Now()
			tc.mockSetup(mock, now)
			user, err := FindByID(db, testUserID)

			if tc.expectNil {
				assert.Nil(t, user)
			} else {
				assert.NotNil(t, user)
			}

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreate_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		findSetup func(mock sqlmock.Sqlmock)
		saveSetup func(mock sqlmock.Sqlmock, now time.Time)
		expectErr string
		expectNil bool
	}{
		{
			name: "success",
			findSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs("new@user.com").
					WillReturnError(sql.ErrNoRows)
			},
			saveSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
					WithArgs("newuser", "new@user.com", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(123, now, now))
			},
			expectErr: "",
			expectNil: false,
		},
		{
			name: "already exists",
			findSetup: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs("exists@user.com").
					WillReturnRows(userRow(1, "exists", "exists@user.com", "hash", true, now, now))
			},
			saveSetup: nil,
			expectErr: "already exists",
			expectNil: true,
		},
		{
			name: "db error on find",
			findSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs("fail@user.com").
					WillReturnError(sql.ErrConnDone)
			},
			saveSetup: nil,
			expectErr: "database error checking for existing email",
			expectNil: true,
		},
		{
			name: "save error",
			findSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
					WithArgs("savefail@user.com").
					WillReturnError(sql.ErrNoRows)
			},
			saveSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
					WithArgs("savefail", "savefail@user.com", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expectErr: "failed to save new user",
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			now := time.Now()
			tc.findSetup(mock)

			if tc.saveSetup != nil {
				tc.saveSetup(mock, now)
			}

			var (
				username, email, password string
			)

			switch tc.name {
			case "success":
				username, email, password = "newuser", "new@user.com", "plainpass"
			case "already exists":
				username, email, password = "exists", "exists@user.com", "plainpass"
			case "db error on find":
				username, email, password = "fail", "fail@user.com", "plainpass"
			case "save error":
				username, email, password = "savefail", "savefail@user.com", "plainpass"
			}

			user, err := Create(db, username, email, password)

			if tc.expectNil {
				assert.Nil(t, user)
			} else {
				assert.NotNil(t, user)
			}

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
