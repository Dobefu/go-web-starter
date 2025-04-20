package user

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

type mockDatabaseWithQueryRowError struct {
	mockDatabase
	queryRowError error
}

func (m *mockDatabaseWithQueryRowError) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func (m *mockDatabaseWithQueryRowError) Stats() sql.DBStats {
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

func TestUserGetID(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, testUserID, user.GetID())
}

func TestUserGetUsername(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, testUsername, user.GetUsername())
}

func TestUserGetEmail(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, testEmail, user.GetEmail())
}

func TestUserGetStatus(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, true, user.GetStatus())
}

func TestUserGetCreatedAt(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, time.Unix(testCreatedAtUnix, 0), user.GetCreatedAt())
}

func TestUserGetUpdatedAt(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, time.Unix(testUpdatedAtUnix, 0), user.GetUpdatedAt())
}

func TestUserSaveQueryRowError(t *testing.T) {
	t.Parallel()

	expectedError := sql.ErrConnDone

	db := &mockDatabaseWithQueryRowError{
		queryRowError: expectedError,
	}

	user := setupUserTests()
	err := user.Save(db)
	assert.EqualError(t, err, expectedError.Error())
}

func TestUserSaveErrScan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	defer func() { _ = mockDB.Close() }()

	db := &mockDatabase{
		mock: mock,
		db:   mockDB,
	}

	user := setupUserTests()

	mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
		WithArgs(user.id, user.username, user.email, user.status, user.createdAt, user.updatedAt).
		WillReturnError(sql.ErrNoRows)

	err = user.Save(db)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestUserSaveSuccess(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	defer func() { _ = mockDB.Close() }()

	db := &mockDatabase{
		mock: mock,
		db:   mockDB,
	}

	user := setupUserTests()

	mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery)).
		WithArgs(user.id, user.username, user.email, user.status, user.createdAt, user.updatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = user.Save(db)
	assert.NoError(t, err)
}
