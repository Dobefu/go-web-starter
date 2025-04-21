package cmd

import (
	"runtime"
	"testing"
	"time"

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

	server.DefaultNew = func(port int) (server.ServerInterface, error) {
		return mockSrv, nil
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
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	var isExitCalled bool

	osExit = func(code int) {
		isExitCalled = true
		assert.Equal(t, 1, code, "Expected exit code 1")
	}

	mockSrv := setupMockServer(t, assert.AnError)

	cmd := &cobra.Command{}
	cmd.Flags().Int("port", defaultTestPort, "")

	ServerCmd(cmd, []string{})

	mockSrv.AssertExpectations(t)
	assert.True(t, isExitCalled, "osExit should have been called")
}

func TestServerCmdInvalidPort(t *testing.T) {
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()
	var isExitCalled bool

	osExit = func(code int) {
		isExitCalled = true
		assert.Equal(t, 1, code, "Expected exit code 1")
		runtime.Goexit()
	}

	mockSrv := setupMockServer(t, nil)

	cmd := &cobra.Command{}

	done := make(chan bool)

	go func() {
		ServerCmd(cmd, []string{})
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)

	mockSrv.AssertNumberOfCalls(t, "Start", 0)
	assert.True(t, isExitCalled, "osExit should have been called due to invalid port flag")
}
