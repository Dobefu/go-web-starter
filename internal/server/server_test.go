package server

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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

func TestNew(t *testing.T) {
	originalDefaultNew := DefaultNew
	defer func() { DefaultNew = originalDefaultNew }()

	mockServer := &Server{port: 8080}

	DefaultNew = func(port int) ServerInterface {
		return mockServer
	}

	server := New(8080)

	assert.Equal(t, mockServer, server)
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

	originalNew := database.New
	database.New = func(cfg config.Database, log *logger.Logger) (*database.Database, error) {
		return &database.Database{}, nil
	}
	defer func() { database.New = originalNew }()

	port := 8080
	srv := defaultNew(port)

	assert.NotNil(t, srv)
	serverImpl, ok := srv.(*Server)
	assert.True(t, ok)
	assert.Equal(t, port, serverImpl.port)
	assert.NotNil(t, serverImpl.router)
	assert.NotNil(t, serverImpl.db)
}

func TestDefaultNewErrors(t *testing.T) {
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	gin.SetMode(gin.ReleaseMode)

	err := os.MkdirAll("templates", 0755)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll("templates") }()

	err = os.WriteFile("templates/index.html", []byte("{{define \"index\"}}test{{end}}"), 0644)
	assert.NoError(t, err)

	err = os.MkdirAll("static", 0755)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll("static") }()

	originalNew := database.New
	defer func() { database.New = originalNew }()

	database.New = func(cfg config.Database, log *logger.Logger) (*database.Database, error) {
		return nil, fmt.Errorf("database error")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic from database error")
		}
	}()

	_ = defaultNew(8080)
}

func TestDefaultNewTemplateError(t *testing.T) {
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	gin.SetMode(gin.TestMode)

	_ = os.RemoveAll("templates")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic from template error")
		}
	}()

	_ = defaultNew(8080)
}

func TestGetDatabaseConfig(t *testing.T) {
	viper.Set("database.host", "localhost")
	viper.Set("database.port", 5432)
	viper.Set("database.user", "testuser")
	viper.Set("database.password", "testpass")
	viper.Set("database.dbname", "testdb")

	config := getDatabaseConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "testuser", config.User)
	assert.Equal(t, "testpass", config.Password)
	assert.Equal(t, "testdb", config.DBName)
}

func TestLoadTemplates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "templates")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	assert.NoError(t, err)

	files := []string{
		filepath.Join(tempDir, "index.html"),
		filepath.Join(subDir, "about.html"),
	}

	for _, file := range files {
		err = os.WriteFile(file, []byte("{{define \"test\"}}test{{end}}"), 0644)
		assert.NoError(t, err)
	}

	templates, err := loadTemplates(tempDir, 0)
	assert.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Contains(t, templates, files[0])
	assert.Contains(t, templates, files[1])

	_, err = loadTemplates(tempDir, 11)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max recursion depth of 10 exceeded")

	_, err = loadTemplates("/non/existent/path", 0)
	assert.Error(t, err)

	err = os.RemoveAll(files[0])
	assert.NoError(t, err)
	err = os.Symlink("/non/existent/target", files[0])
	assert.NoError(t, err)
	_, err = loadTemplates(tempDir, 0)
	assert.Error(t, err)

	deepDir := tempDir

	for i := 0; i < 11; i++ {
		deepDir = filepath.Join(deepDir, fmt.Sprintf("level%d", i))
		err = os.MkdirAll(deepDir, 0755)
		assert.NoError(t, err)
	}

	templateFile := filepath.Join(deepDir, "template.html")
	err = os.WriteFile(templateFile, []byte("{{define \"test\"}}test{{end}}"), 0644)
	assert.NoError(t, err)

	_, err = loadTemplates(tempDir, 0)
	assert.Error(t, err)

	recursiveDir, err := os.MkdirTemp("", "templates3")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(recursiveDir) }()

	subDir3 := filepath.Join(recursiveDir, "subdir3")
	err = os.MkdirAll(subDir3, 0755)
	assert.NoError(t, err)

	brokenLink := filepath.Join(subDir3, "broken.html")
	err = os.Symlink("/non/existent/target", brokenLink)
	assert.NoError(t, err)

	_, err = loadTemplates(recursiveDir, 0)
	assert.Error(t, err)

	finalDir, err := os.MkdirTemp("", "templates4")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(finalDir) }()

	brokenLink = filepath.Join(finalDir, "broken.html")
	err = os.Symlink("/non/existent/target", brokenLink)
	assert.NoError(t, err)

	_, err = loadTemplates(finalDir, 0)
	assert.Error(t, err)
}

func TestStart(t *testing.T) {
	port := 8080
	srv := newTestServer(port)

	assert.NotNil(t, srv)
	err := srv.Start()
	assert.NoError(t, err)
}
