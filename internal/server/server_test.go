package server

import (
	"fmt"
	"os"
	"testing"

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

func TestNewSuccess(t *testing.T) {
	port := 8080
	srv := NewTestServer(port)

	assert.NotNil(t, srv)
}

func TestDefaultNew(t *testing.T) {
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	err := os.MkdirAll("templates", 0755)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll("templates") }()

	err = os.WriteFile("templates/index.html", []byte("{{define \"index\"}}test{{end}}"), 0644)
	assert.NoError(t, err)

	err = os.MkdirAll("static", 0755)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll("static") }()

	port := 8080
	srv := DefaultNew(port).(*Server)

	assert.NotNil(t, srv)
	assert.Equal(t, port, srv.port)
	assert.NotNil(t, srv.router)

	assert.Equal(t, gin.ReleaseMode, gin.Mode())

	wrapper, ok := srv.router.(*routerWrapper)
	assert.True(t, ok)
	assert.NotNil(t, wrapper.Router)
	assert.NotNil(t, wrapper.IRouter)

	engine, ok := wrapper.IRouter.(*gin.Engine)
	assert.True(t, ok)
	assert.NotNil(t, engine)

	handlers := engine.Handlers
	assert.GreaterOrEqual(t, len(handlers), 2)
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	wrapper := &routerWrapper{
		Router:  router,
		IRouter: router,
	}

	srv := &Server{
		router: wrapper,
		port:   8080,
	}

	assert.NotPanics(t, func() {
		srv.registerRoutes()
	})

	engine := wrapper.IRouter.(*gin.Engine)
	routes := engine.Routes()
	assert.Greater(t, len(routes), 0)
}

func TestNew(t *testing.T) {
	originalDefaultNew := DefaultNew
	defer func() { DefaultNew = originalDefaultNew }()

	port := 8080
	DefaultNew = func(p int) ServerInterface {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		srv := &Server{
			router: &routerWrapper{
				Router:  router,
				IRouter: router,
			},
			port: p,
		}
		return srv
	}

	srv := New(port)
	assert.NotNil(t, srv)

	serverImpl, ok := srv.(*Server)
	assert.True(t, ok)
	assert.Equal(t, port, serverImpl.port)
}

func TestStartSuccess(t *testing.T) {
	mockRouter := new(MockRouter)
	mockRouter.On("Run", ":8080").Return(nil)

	srv := &Server{
		router: mockRouter,
		port:   8080,
	}

	err := srv.Start()
	assert.NoError(t, err)
	mockRouter.AssertExpectations(t)
}

func TestStartError(t *testing.T) {
	mockRouter := new(MockRouter)
	expectedErr := fmt.Errorf("router error")
	mockRouter.On("Run", ":8080").Return(expectedErr)

	srv := &Server{
		router: mockRouter,
		port:   8080,
	}

	err := srv.Start()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRouter.AssertExpectations(t)
}
