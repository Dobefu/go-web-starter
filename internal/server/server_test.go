package server

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRouter struct {
	mock.Mock
}

func (m *MockRouter) Run(addr ...string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func TestNewSuccess(t *testing.T) {
	port := 8080
	srv := New(port)

	assert.NotNil(t, srv)
}

func TestDefaultNew(t *testing.T) {
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

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

func TestStartSuccess(t *testing.T) {
	port := 8080
	mockRouter := new(MockRouter)
	srv := &Server{
		router: mockRouter,
		port:   port,
	}

	expectedAddr := fmt.Sprintf(":%d", port)
	mockRouter.On("Run", []string{expectedAddr}).Return(nil)

	err := srv.Start()
	assert.NoError(t, err)
	mockRouter.AssertExpectations(t)
}

func TestStartError(t *testing.T) {
	port := 8080
	mockRouter := new(MockRouter)
	srv := &Server{
		router: mockRouter,
		port:   port,
	}

	expectedAddr := fmt.Sprintf(":%d", port)
	expectedErr := fmt.Errorf("failed to start server")
	mockRouter.On("Run", []string{expectedAddr}).Return(expectedErr)

	err := srv.Start()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRouter.AssertExpectations(t)
}
