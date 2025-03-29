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
	t.Parallel()

	port := 8080
	srv := New(port)

	assert.NotNil(t, srv)
	assert.NotNil(t, srv.router)
	assert.Equal(t, port, srv.port)

	router := srv.router.(*routerWrapper).IRouter.(*gin.Engine)
	handlers := router.Handlers
	assert.Greater(t, len(handlers), 0)
}

func TestStartSuccess(t *testing.T) {
	t.Parallel()

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
