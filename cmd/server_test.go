package cmd

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/server"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const defaultTestPort = 4000

type MockServer struct {
	mock.Mock
}

func (m *MockServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func setupMockServer(t *testing.T, returnError error) *MockServer {
	mockSrv := new(MockServer)
	mockSrv.On("Start").Return(returnError)

	originalNew := server.DefaultNew
	server.DefaultNew = func(port int) server.ServerInterface {
		return mockSrv
	}
	t.Cleanup(func() { server.DefaultNew = originalNew })

	return mockSrv
}

func TestServerCmdSuccess(t *testing.T) {
	mockSrv := setupMockServer(t, nil)

	cmd := &cobra.Command{}
	cmd.Flags().Int("port", defaultTestPort, "")

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})

	mockSrv.AssertExpectations(t)
}

func TestServerCmdError(t *testing.T) {
	mockSrv := setupMockServer(t, assert.AnError)

	cmd := &cobra.Command{}
	cmd.Flags().Int("port", defaultTestPort, "")

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})

	mockSrv.AssertExpectations(t)
}

func TestServerCmdInvalidPort(t *testing.T) {
	mockSrv := setupMockServer(t, nil)

	cmd := &cobra.Command{}

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})

	mockSrv.AssertNumberOfCalls(t, "Start", 0)
}
