package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Query(query string, args ...any) (*sql.Rows, error) {
	mockArgs := m.Called(query, args)

	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}

	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDatabase) QueryRow(query string, args ...any) (*sql.Row, error) {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Row), mockArgs.Error(1)
}

func (m *MockDatabase) Exec(query string, args ...any) (sql.Result, error) {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockDatabase) Begin() (*sql.Tx, error) {
	mockArgs := m.Called()
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Tx), mockArgs.Error(1)
}

func TestHealthCheckSuccess(t *testing.T) {
	t.Parallel()

	mockDB := new(MockDatabase)
	mockDB.On("Ping").Return(nil)

	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("db", mockDB)
		c.Next()
	})

	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])

	mockDB.AssertExpectations(t)
}

func TestHealthCheckDatabaseError(t *testing.T) {
	t.Parallel()

	mockDB := new(MockDatabase)
	mockDB.On("Ping").Return(assert.AnError)

	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("db", mockDB)
		c.Next()
	})

	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.NotNil(t, response["error"])

	mockDB.AssertExpectations(t)
}

func TestHealthCheckNoDatabase(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Database connection not found", response["error"])
}
