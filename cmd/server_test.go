package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockServer struct {
	mock.Mock
}

func (m *MockServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func TestServerCmdSuccess(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Int("port", 4000, "")

	assert.NotPanics(t, func() {
		ServerCmd(cmd, []string{})
	})
}
