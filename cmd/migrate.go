package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configFileNameDefault = "config.toml"
	configPathDefault     = "."
	logLevelConfigKey     = "log.level"
	logLevelDefault       = logger.InfoLevel

	errReadingConfig       = "Error reading config file, %s\n"
	errUnmarshallingConfig = "Error unmarshalling config: %v\n"
	errDbConnection        = "Failed to connect to database"
	errDbClose             = "Error closing database connection"
	errInvalidVersionFmt   = "invalid version format: %s. Please provide an integer"

	logMsgRunningUp      = "Running migrations up..."
	logMsgUpSuccess      = "Migrations applied successfully."
	logMsgUpFailed       = "Migration failed"
	logMsgRunningDown    = "Running migration down..."
	logMsgDownSuccess    = "Last migration rolled back successfully."
	logMsgDownFailed     = "Migration rollback failed"
	logMsgRunningVersion = "Migrating to specific version"
	logMsgVersionSuccess = "Successfully migrated to version"
	logMsgVersionFailed  = "Migration to version failed"

	logFieldError   = "error"
	logFieldVersion = "version"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
}

func setupMigrateEnv(cmd *cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error) {
	viper.SetConfigFile(configFileNameDefault)
	viper.AddConfigPath(configPathDefault)
	viper.AutomaticEnv()
	viper.SetDefault(logLevelConfigKey, int(logLevelDefault))

	cfg := config.DefaultConfig

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, errReadingConfig, err)
		return nil, nil, nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, errUnmarshallingConfig, err)
		return nil, nil, nil, err
	}

	log := logger.New(logger.Level(cfg.Log.Level), cmd.ErrOrStderr())

	db, err := database.New(cfg.Database, log)

	if err != nil {
		log.Error(errDbConnection, logger.Fields{logFieldError: err})
		return nil, nil, nil, err
	}

	return &cfg, log, db, nil
}

func closeDBWithLog(db database.DatabaseInterface, log *logger.Logger) {
	closeErr := db.Close()

	if closeErr != nil {
		log.Error(errDbClose, logger.Fields{logFieldError: closeErr})
	}
}

func runMigrateCommand(
	cmd *cobra.Command,
	setupEnv func(*cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error),
	migrateFunc func(cfg config.Database) error,
	runningMsg, successMsg, errorMsg string,
	logFields logger.Fields,
) error {
	cfg, log, db, err := setupEnv(cmd)

	if err != nil {
		return err
	}

	defer closeDBWithLog(db, log)

	log.Info(runningMsg, logFields)
	err = migrateFunc(cfg.Database)

	if err != nil {
		log.Error(errorMsg, mergeFields(logFields, logger.Fields{logFieldError: err}))
		return err
	}

	log.Info(successMsg, logFields)
	return nil
}

func mergeFields(a, b logger.Fields) logger.Fields {
	if len(a) == 0 {
		return b
	}

	if len(b) == 0 {
		return a
	}

	merged := make(logger.Fields, len(a)+len(b))

	for k, v := range a {
		merged[k] = v
	}

	for k, v := range b {
		merged[k] = v
	}

	return merged
}

func runMigrateUp(
	cmd *cobra.Command,
	setupEnv func(*cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error),
	migrateUp func(cfg config.Database) error,
) error {
	return runMigrateCommand(
		cmd, setupEnv, migrateUp,
		logMsgRunningUp, logMsgUpSuccess, logMsgUpFailed, nil,
	)
}

func runMigrateDown(
	cmd *cobra.Command,
	setupEnv func(*cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error),
	migrateDown func(cfg config.Database) error,
) error {
	return runMigrateCommand(
		cmd, setupEnv, migrateDown,
		logMsgRunningDown, logMsgDownSuccess, logMsgDownFailed, nil,
	)
}

func runMigrateVersion(
	cmd *cobra.Command,
	args []string,
	setupEnv func(*cobra.Command) (*config.Config, *logger.Logger, database.DatabaseInterface, error),
	migrateVersion func(cfg config.Database) (int, error),
) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	versionStr := args[0]
	version, err := strconv.Atoi(versionStr)

	if err != nil {
		return fmt.Errorf(errInvalidVersionFmt, versionStr)
	}

	migrateFunc := func(cfg config.Database) error {
		_, err := migrateVersion(cfg)
		return err
	}

	fields := logger.Fields{logFieldVersion: version}

	return runMigrateCommand(
		cmd, setupEnv, migrateFunc,
		logMsgRunningVersion, logMsgVersionSuccess, logMsgVersionFailed, fields,
	)
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all available migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrateUp(cmd, setupMigrateEnv, database.MigrateUp)
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrateDown(cmd, setupMigrateEnv, database.MigrateDown)
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version [version]",
	Short: "Migrate to a specific version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrateVersion(cmd, args, setupMigrateEnv, database.MigrateVersion)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
}
