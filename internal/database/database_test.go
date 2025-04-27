package database

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	stats    sql.DBStats
	pingErr  error
	queryErr error
}

func (m *mockDB) Close() error {
	return nil
}

func (m *mockDB) Ping() error {
	return m.pingErr
}

func (m *mockDB) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, m.queryErr
}

func (m *mockDB) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func (m *mockDB) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *mockDB) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockDB) Stats() sql.DBStats {
	return m.stats
}

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Database
		wantErr bool
	}{
		{
			name: "valid configuration",
			cfg: config.Database{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			wantErr: false,
		},
		{
			name: "invalid configuration",
			cfg: config.Database{
				Host:     "invalid-host",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))

			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}

			defer func() { _ = db.Close() }()

			database := &Database{
				db:     db,
				logger: logger.New(logger.DebugLevel, os.Stdout),
			}

			if tt.wantErr {
				mock.ExpectPing().WillReturnError(errors.New("connection error"))
			} else {
				mock.ExpectPing()
			}

			err = database.Ping()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDatabaseNotInitialized(t *testing.T) {
	t.Parallel()

	db := &Database{
		db: nil,
	}

	err := db.Ping()

	if err != errNotInitialized {
		t.Errorf("Ping() error = %v, want %v", err, errNotInitialized)
	}

	rows, err := db.Query("SELECT 1")

	if err != errNotInitialized {
		t.Errorf("Query() error = %v, want %v", err, errNotInitialized)
	}

	if rows != nil {
		t.Error("Query() rows should be nil when database is not initialized")
	}

	row := db.QueryRow("SELECT 1")

	if row != nil {
		t.Error("QueryRow() should return nil when database is not initialized")
	}

	result, err := db.Exec("INSERT INTO test (name) VALUES (?)", "test")

	if err != errNotInitialized {
		t.Errorf("Exec() error = %v, want %v", err, errNotInitialized)
	}

	if result != nil {
		t.Error("Exec() result should be nil when database is not initialized")
	}

	tx, err := db.Begin()

	if err != errNotInitialized {
		t.Errorf("Begin() error = %v, want %v", err, errNotInitialized)
	}

	if tx != nil {
		t.Error("Begin() transaction should be nil when database is not initialized")
	}

	err = db.Close()

	if err != errNotInitialized {
		t.Errorf("Close() error = %v, want %v", err, errNotInitialized)
	}

	stats := db.Stats()

	if stats != (sql.DBStats{}) {
		t.Error("Stats() should return empty DBStats when database is not initialized")
	}
}

func setupTestDB(t *testing.T) (*Database, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	database := &Database{
		db:     db,
		logger: logger.New(logger.DebugLevel, os.Stdout),
	}

	return database, mock
}

func TestDatabasePing(t *testing.T) {
	t.Parallel()

	database, mock := setupTestDB(t)
	defer func() { _ = database.Close() }()

	mock.ExpectPing()
	err := database.Ping()
	assert.NoError(t, err)

	mock.ExpectPing().WillReturnError(errors.New("ping failed"))
	err = database.Ping()
	assert.Error(t, err)
	assert.Equal(t, "ping failed", err.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseExec(t *testing.T) {
	t.Parallel()

	database, mock := setupTestDB(t)
	defer func() { _ = database.Close() }()

	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(1, 1))
	result, err := database.Exec("CREATE TABLE test")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mock.ExpectExec("DROP TABLE test").WillReturnError(errors.New("exec failed"))
	result, err = database.Exec("DROP TABLE test")
	assert.Error(t, err)
	assert.Equal(t, "exec failed", err.Error())
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBasicDatabaseOperations(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer func() { _ = db.Close() }()

	database := &Database{
		db:     db,
		logger: logger.New(logger.DebugLevel, os.Stdout),
	}

	t.Run("Query", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

		result, err := database.Query("SELECT 1")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		err = result.Close()
		assert.NoError(t, err)
	})

	t.Run("QueryRow", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

		row := database.QueryRow("SELECT 1")
		assert.NotNil(t, row)

		var id int

		err := row.Scan(&id)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	t.Run("Exec", func(t *testing.T) {
		mock.ExpectExec("CREATE").WillReturnResult(sqlmock.NewResult(1, 1))

		result, err := database.Exec("CREATE TABLE test")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Begin", func(t *testing.T) {
		mock.ExpectBegin()

		tx, err := database.Begin()
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})

	t.Run("Ping", func(t *testing.T) {
		mock.ExpectPing()

		err := database.Ping()
		assert.NoError(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseLogging(t *testing.T) {
	t.Parallel()

	mockDB := &mockDB{
		stats: sql.DBStats{
			MaxIdleClosed:      10,
			MaxOpenConnections: 100,
			MaxLifetimeClosed:  int64(time.Hour),
		},
	}

	log := logger.New(logger.DebugLevel, os.Stdout)

	db := &Database{
		db:     mockDB,
		logger: log,
	}

	_, err := db.Query("SELECT 1")
	assert.NoError(t, err)

	mockDB.queryErr = errors.New("query failed")
	_, err = db.Query("INVALID SQL")
	assert.Error(t, err)
	assert.Equal(t, "query failed", err.Error())
}

func TestDatabaseConnectionPool(t *testing.T) {
	t.Parallel()

	mockDB := &mockDB{
		stats: sql.DBStats{
			MaxIdleClosed:      10,
			MaxOpenConnections: 100,
			MaxLifetimeClosed:  int64(time.Hour),
		},
	}

	log := logger.New(logger.DebugLevel, os.Stdout)

	db := &Database{
		db:     mockDB,
		logger: log,
	}

	assert.Equal(t, int64(10), int64(db.Stats().MaxIdleClosed))
	assert.Equal(t, int64(100), int64(db.Stats().MaxOpenConnections))
	assert.Equal(t, int64(time.Hour), db.Stats().MaxLifetimeClosed)
}

func TestQuerySuccess(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test")
	mock.ExpectQuery("SELECT id, name FROM test").WillReturnRows(rows)

	database := &Database{
		db:     db,
		logger: logger.New(logger.DebugLevel, os.Stdout),
	}

	resultRows, err := database.Query("SELECT id, name FROM test")
	assert.NoError(t, err)
	assert.NotNil(t, resultRows)
	err = resultRows.Close()
	assert.NoError(t, err)
}

func TestQueryFailure(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT id, name FROM test").WillReturnError(errors.New("query failed"))

	database := &Database{
		db:     db,
		logger: logger.New(logger.DebugLevel, os.Stdout),
	}

	resultRows, err := database.Query("SELECT id, name FROM test")
	assert.Error(t, err)
	assert.Nil(t, resultRows)
}
