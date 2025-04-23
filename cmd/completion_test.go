package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunCompletionCmd(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		wantErr   bool
	}{
		{
			name:      "bash completion",
			shellType: "bash",
			wantErr:   false,
		},
		{
			name:      "zsh completion",
			shellType: "zsh",
			wantErr:   false,
		},
		{
			name:      "fish completion",
			shellType: "fish",
			wantErr:   false,
		},
		{
			name:      "powershell completion",
			shellType: "powershell",
			wantErr:   false,
		},
		{
			name:      "unsupported shell type",
			shellType: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{
				Use: "test",
			}

			rootCmd.AddCommand(completionCmd)

			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			assert.NoError(t, err)
			os.Stdout = w

			err = runCompletionCmd(rootCmd, []string{tt.shellType})

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported shell type")
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, buf.String(), "completion output should not be empty")
			}
		})
	}
}
