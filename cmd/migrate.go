package cmd

import (
	"fmt"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DatabaseMigrator interface {
	MigrateUp(cfg config.Database) error
	MigrateDown(cfg config.Database) error
	MigrateVersion(cfg config.Database) (int, error)
}

type defaultDatabaseMigrator struct{}

func (d *defaultDatabaseMigrator) MigrateUp(cfg config.Database) error {
	return database.MigrateUp(cfg)
}

func (d *defaultDatabaseMigrator) MigrateDown(cfg config.Database) error {
	return database.MigrateDown(cfg)
}

func (d *defaultDatabaseMigrator) MigrateVersion(cfg config.Database) (int, error) {
	return database.MigrateVersion(cfg)
}

var migrator DatabaseMigrator = &defaultDatabaseMigrator{}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Run:   migrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Run:   migrateDown,
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	Run:   migrateVersion,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
}

func migrateUp(cmd *cobra.Command, args []string) {
	err := migrator.MigrateUp(getDatabaseConfig())

	if err != nil {
		panic(err)
	}
}

func migrateDown(cmd *cobra.Command, args []string) {
	err := migrator.MigrateDown(getDatabaseConfig())

	if err != nil {
		panic(err)
	}
}

func migrateVersion(cmd *cobra.Command, args []string) {
	version, err := migrator.MigrateVersion(getDatabaseConfig())

	if err != nil {
		panic(err)
	}

	fmt.Printf("The migrations are at version %d\n", version)
}

func getDatabaseConfig() config.Database {
	return config.Database{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
	}
}
