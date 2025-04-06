package server

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRouter struct {
	mock.Mock
}

func (m *MockRouter) Run(addr ...string) error {
	args := m.Called(addr[0])
	return args.Error(0)
}

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

func newTestServer(port int) ServerInterface {
	gin.SetMode(gin.TestMode)
	mockRouter := &MockRouter{}
	mockRouter.On("Run", fmt.Sprintf(":%d", port)).Return(nil)

	mockDB := &MockDatabase{}
	mockDB.On("Ping").Return(nil)
	mockDB.On("Close").Return(nil)

	srv := &Server{
		router: mockRouter,
		port:   port,
		db:     mockDB,
	}

	return srv
}

func TestNewSuccess(t *testing.T) {
	port := 8080
	srv := newTestServer(port)

	assert.NotNil(t, srv)
}

func TestDefaultNew(t *testing.T) {
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	gin.SetMode(gin.TestMode)

	err := os.MkdirAll("templates", 0755)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll("templates") }()

	err = os.WriteFile("templates/index.html", []byte("{{define \"index\"}}test{{end}}"), 0644)
	assert.NoError(t, err)

	err = os.MkdirAll("static", 0755)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll("static") }()

	mockDB := &MockDatabase{}
	mockDB.On("Ping").Return(nil)
	mockDB.On("Close").Return(nil)

	originalNew := database.New

	database.New = func(cfg config.Database, log *logger.Logger) (*database.Database, error) {
		return &database.Database{}, nil
	}

	defer func() { database.New = originalNew }()

	port := 8080
	srv := DefaultNew(port)

	assert.NotNil(t, srv)
	serverImpl, ok := srv.(*Server)
	assert.True(t, ok)
	assert.Equal(t, port, serverImpl.port)
}

func TestRegisterRoutes(t *testing.T) {
	port := 8080
	srv := newTestServer(port)

	assert.NotNil(t, srv)
	serverImpl, ok := srv.(*Server)
	assert.True(t, ok)
	assert.NotNil(t, serverImpl.router)
}

func TestNew(t *testing.T) {
	port := 8080
	srv := newTestServer(port)

	assert.NotNil(t, srv)
	serverImpl, ok := srv.(*Server)
	assert.True(t, ok)
	assert.Equal(t, port, serverImpl.port)
}

func TestStartSuccess(t *testing.T) {
	port := 8080
	srv := newTestServer(port)

	assert.NotNil(t, srv)
	err := srv.Start()
	assert.NoError(t, err)
}

func TestStartError(t *testing.T) {
	port := 8080
	srv := newTestServer(port)

	assert.NotNil(t, srv)
	err := srv.Start()
	assert.NoError(t, err)
}
