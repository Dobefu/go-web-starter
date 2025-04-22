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

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all available migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile(configFileNameDefault)
		viper.AddConfigPath(configPathDefault)
		viper.AutomaticEnv()
		viper.SetDefault(logLevelConfigKey, int(logLevelDefault))

		cfg := config.DefaultConfig

		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, errReadingConfig, err)
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, errUnmarshallingConfig, err)
			return err
		}

		log := logger.New(logger.Level(cfg.Log.Level), cmd.ErrOrStderr())

		dbCfg := getDatabaseConfig(&cfg)
		db, err := database.New(dbCfg, log)

		if err != nil {
			log.Error(errDbConnection, logger.Fields{logFieldError: err})
			return err
		}

		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				log.Error(errDbClose, logger.Fields{logFieldError: closeErr})
			}
		}()

		log.Info(logMsgRunningUp, nil)
		err = database.MigrateUp(dbCfg)

		if err != nil {
			log.Error(logMsgUpFailed, logger.Fields{logFieldError: err})
			return err
		}

		log.Info(logMsgUpSuccess, nil)
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile(configFileNameDefault)
		viper.AddConfigPath(configPathDefault)
		viper.AutomaticEnv()
		viper.SetDefault(logLevelConfigKey, int(logLevelDefault))

		cfg := config.DefaultConfig

		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, errReadingConfig, err)
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, errUnmarshallingConfig, err)
			return err
		}

		log := logger.New(logger.Level(cfg.Log.Level), cmd.ErrOrStderr())

		dbCfg := getDatabaseConfig(&cfg)
		db, err := database.New(dbCfg, log)

		if err != nil {
			log.Error(errDbConnection, logger.Fields{logFieldError: err})
			return err
		}

		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				log.Error(errDbClose, logger.Fields{logFieldError: closeErr})
			}
		}()

		log.Info(logMsgRunningDown, nil)
		err = database.MigrateDown(dbCfg)

		if err != nil {
			log.Error(logMsgDownFailed, logger.Fields{logFieldError: err})
			return err
		}

		log.Info(logMsgDownSuccess, nil)
		return nil
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version [version]",
	Short: "Migrate to a specific version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		versionStr := args[0]
		version, err := strconv.Atoi(versionStr)

		if err != nil {
			return fmt.Errorf(errInvalidVersionFmt, versionStr)
		}

		viper.SetConfigFile(configFileNameDefault)
		viper.AddConfigPath(configPathDefault)
		viper.AutomaticEnv()
		viper.SetDefault(logLevelConfigKey, int(logLevelDefault))

		cfg := config.DefaultConfig

		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, errReadingConfig, err)
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, errUnmarshallingConfig, err)
			return err
		}

		log := logger.New(logger.Level(cfg.Log.Level), cmd.ErrOrStderr())

		dbCfg := getDatabaseConfig(&cfg)
		db, err := database.New(dbCfg, log)

		if err != nil {
			log.Error(errDbConnection, logger.Fields{logFieldError: err})
			return err
		}

		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				log.Error(errDbClose, logger.Fields{logFieldError: closeErr})
			}
		}()

		log.Info(logMsgRunningVersion, logger.Fields{logFieldVersion: version})
		_, err = database.MigrateVersion(dbCfg)

		if err != nil {
			log.Error(logMsgVersionFailed, logger.Fields{logFieldVersion: version, logFieldError: err})
			return err
		}

		log.Info(logMsgVersionSuccess, logger.Fields{logFieldVersion: version})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
}
func getDatabaseConfig(cfg *config.Config) config.Database {
	return cfg.Database
}
