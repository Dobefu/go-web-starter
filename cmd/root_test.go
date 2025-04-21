package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/pelletier/go-toml/v2"
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

func TestInitConfigVerbose(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	oldVerbose := verbose
	defer func() { verbose = oldVerbose }()
	verbose = 4

	initConfig()

	assert.Equal(t, logger.Level(0), config.GetLogLevel())
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

func TestInitConfigDefaultCreation(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	tmpDir, err := os.MkdirTemp("", "config_test_*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalWd, err := os.Getwd()
	assert.NoError(t, err)

	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	cfgFile = ""

	viper.Reset()
	cobra.OnInitialize(initConfig)

	initConfig()

	createdConfigFile := defaultConfigFileName
	_, err = os.Stat(createdConfigFile)
	assert.NoError(t, err, "config.toml should be created")

	contentBytes, err := os.ReadFile(createdConfigFile)
	assert.NoError(t, err, "should be able to read created config.toml")

	var createdConfig config.Config
	err = toml.Unmarshal(contentBytes, &createdConfig)
	assert.NoError(t, err, "created config.toml should be valid TOML")

	assert.Equal(t, config.DefaultConfig, createdConfig, "created config.toml content should match DefaultConfig")
}

func captureStderr(f func()) (string, error) {
	originalStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = originalStderr

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func TestInitConfigReadError(t *testing.T) {
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	tmpFile, err := os.CreateTemp("", "malformed_config_*.toml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("[invalid toml content = ?")
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	cfgFile = tmpFile.Name()

	viper.Reset()
	cobra.OnInitialize(initConfig)

	stderrOutput, err := captureStderr(initConfig)
	assert.NoError(t, err, "stderr capture should not fail")

	assert.Contains(t, stderrOutput, "Error reading config file", "stderr should contain config reading error")
}
