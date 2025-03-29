package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func setupRootCmdTest(t *testing.T) (*os.File, func()) {
	tmpFile, err := os.CreateTemp("", "test_config_*.toml")
	assert.NoError(t, err)
	return tmpFile, func() { _ = os.Remove(tmpFile.Name()) }
}

func TestInitConfigSuccess(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	tmpFile, cleanup := setupRootCmdTest(t)
	defer cleanup()

	err := os.WriteFile(tmpFile.Name(), []byte("test = true"), 0644)
	assert.NoError(t, err)

	cfgFile = tmpFile.Name()
	initConfig()

	assert.Equal(t, tmpFile.Name(), viper.ConfigFileUsed())
}

func TestRootExecuteSuccess(t *testing.T) {
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	rootCmd = &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	assert.NotPanics(t, Execute)
}

func TestExecuteWithErrorHelper(t *testing.T) {
	if os.Getenv("GO_TEST_EXIT") != "1" {
		return
	}

	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	rootCmd = &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return assert.AnError
		},
	}

	Execute()
}

func TestInitConfigCustomFile(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	tmpFile, cleanup := setupRootCmdTest(t)
	defer cleanup()

	cfgFile = tmpFile.Name()
	initConfig()

	assert.Equal(t, tmpFile.Name(), viper.ConfigFileUsed())
}

func TestExecuteErrFlag(t *testing.T) {
	cmd := exec.Command(os.Args[0], "bogus")
	cmd.Env = append(os.Environ(), "GO_TEST_EXIT=1")
	err := cmd.Run()

	if exitErr, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, exitErr.ExitCode())
	}
}

func TestConfigFlag(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	tmpFile, cleanup := setupRootCmdTest(t)
	defer cleanup()

	rootCmd.SetArgs([]string{"--config", tmpFile.Name()})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), cfgFile)
}
