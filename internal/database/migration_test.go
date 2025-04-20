package database

import (
	"database/sql"
	"errors"
	"io/fs"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/stretchr/testify/assert"
)

type MockFS struct {
	files []fs.DirEntry
	err   error
}

type mockFile struct {
	name string
}

func (m mockFile) Name() string               { return m.name }
func (m mockFile) IsDir() bool                { return false }
func (m mockFile) Type() fs.FileMode          { return 0 }
func (m mockFile) Info() (fs.FileInfo, error) { return nil, nil }

func (m *MockFS) Open(name string) (fs.File, error)          { return nil, m.err }
func (m *MockFS) ReadDir(name string) ([]fs.DirEntry, error) { return m.files, m.err }
func (m *MockFS) ReadFile(name string) ([]byte, error) {
	return []byte("CREATE TABLE test (id int);"), m.err
}

func setupTest(t *testing.T) (sqlmock.Sqlmock, *Database, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mockFS := &MockFS{
		files: []fs.DirEntry{
			mockFile{name: "000001_create_users_table.up.sql"},
			mockFile{name: "000001_create_users_table.down.sql"},
		},
	}

	originalContentFS := contentFS
	contentFS = ContentFS{content: mockFS}

	originalNew := New

	New = func(cfg config.Database, log *logger.Logger) (*Database, error) {
		return &Database{db: db}, nil
	}

	cleanup := func() {
		_ = db.Close()
		contentFS = originalContentFS
		New = originalNew
	}

	return mock, &Database{db: db}, cleanup
}

func TestMigrateUp(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		expectError   bool
		errorContains string
	}{
		{
			name: "First Migration",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnError(sql.ErrNoRows)
				mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("TRUNCATE migrations").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("INSERT INTO migrations").WithArgs(1, false).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Dirty State",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnRows(
					sqlmock.NewRows([]string{"version", "dirty"}).AddRow(1, true),
				)
			},
			expectError:   true,
			errorContains: "migrations table is in a dirty state",
		},
		{
			name: "Create Table Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnError(errors.New("create table error"))
			},
			expectError:   true,
			errorContains: "create table error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _, cleanup := setupTest(t)
			defer cleanup()

			tt.setupMock(mock)

			err := MigrateUp(getTestConfig())

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMigrateDown(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		expectError   bool
		errorContains string
	}{
		{
			name: "Revert From Version 2",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnRows(
					sqlmock.NewRows([]string{"version", "dirty"}).AddRow(2, false),
				)
				mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("TRUNCATE migrations").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("INSERT INTO migrations").WithArgs(0, false).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "No Migrations to Revert",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnRows(
					sqlmock.NewRows([]string{"version", "dirty"}).AddRow(0, false),
				)
			},
		},
		{
			name: "Dirty State",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnRows(
					sqlmock.NewRows([]string{"version", "dirty"}).AddRow(1, true),
				)
			},
			expectError:   true,
			errorContains: "migrations table is in a dirty state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _, cleanup := setupTest(t)
			defer cleanup()

			tt.setupMock(mock)

			err := MigrateDown(getTestConfig())

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMigrationState(t *testing.T) {
	mock, db, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		wantVersion int
		wantDirty   bool
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnRows(
					sqlmock.NewRows([]string{"version", "dirty"}).AddRow(2, false),
				)
			},
			wantVersion: 2,
			wantDirty:   false,
		},
		{
			name: "No Rows",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnError(sql.ErrNoRows)
			},
			wantVersion: 0,
			wantDirty:   false,
		},
		{
			name: "Database Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version,dirty FROM migrations").WillReturnError(errors.New("database error"))
			},
			wantVersion: 0,
			wantDirty:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)
			version, dirty := getMigrationState(db)
			assert.Equal(t, tt.wantVersion, version)
			assert.Equal(t, tt.wantDirty, dirty)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRunMigration(t *testing.T) {
	mock, db, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		mockFS        *MockFS
		expectError   bool
		errorContains string
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name: "Execute Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE test").WillReturnError(errors.New("execute error"))
			},
			expectError:   true,
			errorContains: "execute error",
		},
		{
			name:          "Read File Error",
			mockFS:        &MockFS{err: errors.New("read error")},
			expectError:   true,
			errorContains: "read error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockFS != nil {
				originalContentFS := contentFS
				contentFS = ContentFS{content: tt.mockFS}
				defer func() { contentFS = originalContentFS }()
			}

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			err := runMigration(db, "test.sql", 1)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func getTestConfig() config.Database {
	return config.Database{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		DBName:   "test",
	}
}

func TestDatabaseOperations(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)

	defer func() { _ = db.Close() }()

	database := &Database{db: db}

	tests := []struct {
		name        string
		operation   func() error
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "Query Success",
			operation: func() error {
				result, err := database.Query("SELECT id, name FROM test")

				if err != nil {
					return err
				}

				defer func() { _ = result.Close() }()
				return nil
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name FROM test").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test"),
				)
			},
		},
		{
			name: "Query Error",
			operation: func() error {
				_, err := database.Query("SELECT id, name FROM test")
				return err
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name FROM test").WillReturnError(errors.New("query error"))
			},
			expectError: true,
		},
		{
			name: "QueryRow Success",
			operation: func() error {
				var id int
				row := database.QueryRow("SELECT id FROM test")
				return row.Scan(&id)
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM test").WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			},
		},
		{
			name: "QueryRow Error",
			operation: func() error {
				var id int
				row := database.QueryRow("SELECT id FROM test")
				return row.Scan(&id)
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM test").WillReturnError(errors.New("query error"))
			},
			expectError: true,
		},
		{
			name: "Exec Success",
			operation: func() error {
				_, err := database.Exec("INSERT INTO test VALUES ($1)", "test")
				return err
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test").WithArgs("test").WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "Exec Error",
			operation: func() error {
				_, err := database.Exec("INSERT INTO test VALUES ($1)", "test")
				return err
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test").WithArgs("test").WillReturnError(errors.New("exec error"))
			},
			expectError: true,
		},
		{
			name: "Begin Success",
			operation: func() error {
				tx, err := database.Begin()

				if err != nil {
					return err
				}

				return tx.Rollback()
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Begin Error",
			operation: func() error {
				_, err := database.Begin()
				return err
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectError: true,
		},
		{
			name: "Ping Success",
			operation: func() error {
				return database.Ping()
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
		},
		{
			name: "Ping Error",
			operation: func() error {
				return database.Ping()
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("ping error"))
			},
			expectError: true,
		},
		{
			name: "Close Success",
			operation: func() error {
				return database.Close()
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectClose()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)
			err := tt.operation()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
