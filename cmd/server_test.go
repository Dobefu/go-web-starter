package cmd

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/server"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testPort = 4000

type MockServer struct {
	mock.Mock
}

func (m *MockServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func TestServerCmdSuccess(t *testing.T) {
	mockSrv := new(MockServer)
	mockSrv.On("Start").Return(nil)

	originalNew := server.DefaultNew
	server.DefaultNew = func(port int) server.ServerInterface {
		return mockSrv
	}

	defer func() { server.DefaultNew = originalNew }()

	cmd := &cobra.Command{}
	cmd.Flags().Int("port", testPort, "")

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})

	mockSrv.AssertExpectations(t)
}

func TestServerCmdError(t *testing.T) {
	mockSrv := new(MockServer)
	mockSrv.On("Start").Return(assert.AnError)

	originalNew := server.DefaultNew
	server.DefaultNew = func(port int) server.ServerInterface { return mockSrv }

	defer func() { server.DefaultNew = originalNew }()

	cmd := &cobra.Command{}
	cmd.Flags().Int("port", testPort, "")

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})

	mockSrv.AssertExpectations(t)
}

func TestServerCmdInvalidPort(t *testing.T) {
	mockSrv := new(MockServer)

	originalNew := server.DefaultNew
	server.DefaultNew = func(port int) server.ServerInterface {
		return mockSrv
	}
	defer func() { server.DefaultNew = originalNew }()

	cmd := &cobra.Command{}

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})

	mockSrv.AssertNumberOfCalls(t, "Start", 0)
}
