package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func setupMigrateTest(t *testing.T) func() {
	t.Helper()

	tempDir := t.TempDir()
	dummyConfigPath := filepath.Join(tempDir, "config.toml")
	dummyConfigContent := `
[Database]
  Host = "localhost"
  Port = 54329 # Use a non-standard port to avoid real connections
  User = "testuser"
  Password = "testpassword"
  DBName = "testdb"
[Log]
  Level = 2 # Use integer for log level (e.g., 2 for Info)
`
	err := os.WriteFile(dummyConfigPath, []byte(dummyConfigContent), 0600)
	assert.NoError(t, err, "Failed to write dummy config file")

	originalArgs := os.Args
	originalWd, err := os.Getwd()
	assert.NoError(t, err)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	viper.Reset()

	return func() {
		viper.Reset()
		os.Args = originalArgs
		err := os.Chdir(originalWd)
		assert.NoError(t, err)
	}
}

func executeCommand(args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs(args)
	rootCmd.SilenceUsage = true

	err = rootCmd.Execute()

	rootCmd.SilenceUsage = false

	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestMigrateUpCommand(t *testing.T) {
	cleanup := setupMigrateTest(t)
	defer cleanup()

	stdout, stderr, err := executeCommand("migrate", "up")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
	assert.Contains(t, stderr, fmt.Sprintf(`ERROR "%s"`, errDbConnection))
	assert.Empty(t, stdout)
}

func TestMigrateDownCommand(t *testing.T) {
	cleanup := setupMigrateTest(t)
	defer cleanup()

	stdout, stderr, err := executeCommand("migrate", "down")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
	assert.Contains(t, stderr, fmt.Sprintf(`ERROR "%s"`, errDbConnection))
	assert.Empty(t, stdout)
}

func TestMigrateVersionCommand(t *testing.T) {
	cleanup := setupMigrateTest(t)
	defer cleanup()

	stdout, stderr, err := executeCommand("migrate", "version", "1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
	assert.Contains(t, stderr, fmt.Sprintf(`ERROR "%s"`, errDbConnection))
	assert.Empty(t, stdout)

	stdout, stderr, err = executeCommand("migrate", "version", "abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf(errInvalidVersionFmt, "abc"))
	assert.NotContains(t, stderr, errDbConnection)
	assert.Empty(t, stdout)

	stdout, _, err = executeCommand("migrate", "version")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
	assert.Empty(t, stdout)
}
