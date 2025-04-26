package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"errors"
	"strings"

	"database/sql"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/cobra"
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
  Port = 54329
  User = "testuser"
  Password = "testpassword"
  DBName = "testdb"
[Log]
  Level = 2
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

func TestMigrateConfigFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	viper.Reset()

	stdout, stderr, err := executeCommand("migrate", "up")
	assert.NoError(t, err)

	output := stdout + stderr
	assert.Contains(t, output, "Database connection established")
	assert.Contains(t, output, "Migrations applied successfully.")

	_, statErr := os.Stat("config.toml")
	assert.NoError(t, statErr, "config.toml should be created")
}

func TestMigrateMalformedConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	malformedConfigPath := filepath.Join(tempDir, "config.toml")
	malformedContent := "[invalid toml = ?"

	err := os.WriteFile(malformedConfigPath, []byte(malformedContent), 0600)
	assert.NoError(t, err)

	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	viper.Reset()
	stdout, stderr, err := executeCommand("migrate", "up")
	assert.Error(t, err)
	assert.Contains(t, stderr, "toml: expected character ]")
	assert.Empty(t, stdout)
}

func TestSetupMigrateEnv(t *testing.T) {
	tempDir := t.TempDir()
	validConfig := `[Database]\n  Host = "localhost"\n  Port = 5432\n  User = "testuser"\n  Password = "testpassword"\n  DBName = "testdb"\n[Log]\n  Level = 2\n`
	unmarshalConfig := `[Log]\n  Level = "not_an_int"\n`

	tests := []struct {
		name      string
		setup     func(configFilePath string)
		dbErr     error
		expectErr string
	}{
		{
			name:  "config file not found",
			setup: func(configFilePath string) {},
		},
		{
			name: "config file read error",
			setup: func(configFilePath string) {
				viper.SetConfigFile(filepath.Join(configFilePath, "doesnotexist.toml"))
			},
			expectErr: "no such file or directory",
		},
		{
			name: "unmarshal error",
			setup: func(configFilePath string) {
				_ = os.WriteFile(configFilePath, []byte(unmarshalConfig), 0600)
			},
			expectErr: "While parsing config: toml:",
		},
		{
			name: "db connection error",
			setup: func(configFilePath string) {
				_ = os.WriteFile(configFilePath, []byte(validConfig), 0600)
			},
			dbErr:     errors.New("db error"),
			expectErr: "db error",
		},
		{
			name: "success",
			setup: func(configFilePath string) {
				_ = os.WriteFile(configFilePath, []byte(validConfig), 0600)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("VIPER_CONFIG_PATH", tempDir)
			configPath := tempDir
			configFilePath := filepath.Join(configPath, "config.toml")
			_ = os.Remove(configFilePath)

			if tt.setup != nil {
				tt.setup(configFilePath)
			}

			origDBNew := database.New
			database.New = func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
				if tt.dbErr != nil {
					return nil, tt.dbErr
				}

				return &mockDB{}, nil
			}

			defer func() { database.New = origDBNew }()

			stderr := new(strings.Builder)
			cmd := &cobra.Command{}
			cmd.SetErr(stderr)

			viper.Reset()
			viper.SetConfigFile(configFilePath)
			viper.AddConfigPath(configPath)
			viper.AutomaticEnv()

			cfg, logg, db, err := setupMigrateEnv(cmd)

			if tt.name == "db connection error" {
				if err == nil {
					t.Fatalf("expected error for db connection error, got nil")
				}

				if !strings.Contains(err.Error(), "db error") {
					t.Errorf("expected error to contain 'db error', got %q", err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error for %s, got: %v", tt.name, err)
			}

			if cfg == nil || logg == nil || db == nil {
				t.Fatalf("expected non-nil cfg, log, db")
			}
		})
	}
}

func migrationTestDeps(
	setupEnvErr error,
	closeErr error,
) (
	setupEnv func(*cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error),
	logOutput *strings.Builder,
) {
	logOutput = new(strings.Builder)
	fakeLog := logger.New(logger.InfoLevel, logOutput)
	fakeDB := &mockDBClose{closeErr: closeErr}
	fakeCfg := &config.Config{Database: config.Database{}}

	setupEnv = func(*cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error) {
		if setupEnvErr != nil {
			return nil, nil, nil, setupEnvErr
		}

		return fakeCfg, fakeLog, fakeDB, nil
	}

	return
}

func assertMigrationTestResult(t *testing.T, err error, expectErr bool, expectLog string, logOutput *strings.Builder, closeErr error, expectErrMsg string) {
	t.Helper()

	if expectErr {
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if expectErrMsg != "" && !strings.Contains(err.Error(), expectErrMsg) {
			t.Errorf("expected error message to contain %q, got %q", expectErrMsg, err.Error())
		}

		if expectLog != "" && !strings.Contains(logOutput.String(), expectLog) {
			t.Errorf("expected log to contain %q, got %q", expectLog, logOutput.String())
		}
	} else {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if closeErr != nil && !strings.Contains(logOutput.String(), "Error closing database connection") {
		t.Errorf("expected log to contain db close error, got %q", logOutput.String())
	}
}

func TestMigrateUpCmd_RunE(t *testing.T) {
	tests := []struct {
		name         string
		setupEnvErr  error
		migrateUpErr error
		closeErr     error
		expectErr    bool
		expectLog    string
	}{
		{"success", nil, nil, nil, false, ""},
		{"migration error", nil, errors.New("migration failed"), nil, true, "Migration failed"},
		{"db close error", nil, nil, errors.New("close failed"), false, ""},
		{"setup env error", errors.New("env error"), nil, nil, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv, logOutput := migrationTestDeps(tt.setupEnvErr, tt.closeErr)
			migrateUp := func(cfg config.Database) error { return tt.migrateUpErr }
			cmd := &cobra.Command{}
			err := runMigrateUp(cmd, setupEnv, migrateUp)
			assertMigrationTestResult(t, err, tt.expectErr, tt.expectLog, logOutput, tt.closeErr, "")
		})
	}
}

func TestMigrateDownCmd_RunE(t *testing.T) {
	tests := []struct {
		name           string
		setupEnvErr    error
		migrateDownErr error
		closeErr       error
		expectErr      bool
		expectLog      string
	}{
		{"success", nil, nil, nil, false, ""},
		{"migration error", nil, errors.New("down failed"), nil, true, "Migration rollback failed"},
		{"db close error", nil, nil, errors.New("close failed"), false, ""},
		{"setup env error", errors.New("env error"), nil, nil, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv, logOutput := migrationTestDeps(tt.setupEnvErr, tt.closeErr)
			migrateDown := func(cfg config.Database) error { return tt.migrateDownErr }
			cmd := &cobra.Command{}
			err := runMigrateDown(cmd, setupEnv, migrateDown)
			assertMigrationTestResult(t, err, tt.expectErr, tt.expectLog, logOutput, tt.closeErr, "")
		})
	}
}

func TestMigrateVersionCmd_RunE(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		setupEnvErr       error
		migrateVersionErr error
		closeErr          error
		expectErr         bool
		expectLog         string
		expectErrMsg      string
	}{
		{"success", []string{"1"}, nil, nil, nil, false, "", ""},
		{"migration error", []string{"2"}, nil, errors.New("version failed"), nil, true, "Migration to version failed", ""},
		{"db close error", []string{"3"}, nil, nil, errors.New("close failed"), false, "", ""},
		{"setup env error", []string{"4"}, errors.New("env error"), nil, nil, true, "", ""},
		{"invalid version arg", []string{"abc"}, nil, nil, nil, true, "", "invalid version format: abc. Please provide an integer"},
		{"missing version arg", []string{}, nil, nil, nil, true, "", "accepts 1 arg(s), received 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv, logOutput := migrationTestDeps(tt.setupEnvErr, tt.closeErr)
			migrateVersion := func(cfg config.Database) (int, error) { return 0, tt.migrateVersionErr }
			cmd := &cobra.Command{}
			err := runMigrateVersion(cmd, tt.args, setupEnv, migrateVersion)
			assertMigrationTestResult(t, err, tt.expectErr, tt.expectLog, logOutput, tt.closeErr, tt.expectErrMsg)
		})
	}
}

type mockDBClose struct{ closeErr error }

func (m *mockDBClose) Close() error                                       { return m.closeErr }
func (m *mockDBClose) Ping() error                                        { return nil }
func (m *mockDBClose) Query(query string, args ...any) (*sql.Rows, error) { return nil, nil }
func (m *mockDBClose) QueryRow(query string, args ...any) *sql.Row        { return nil }
func (m *mockDBClose) Exec(query string, args ...any) (sql.Result, error) { return nil, nil }
func (m *mockDBClose) Begin() (*sql.Tx, error)                            { return nil, nil }
func (m *mockDBClose) Stats() sql.DBStats                                 { return sql.DBStats{} }
