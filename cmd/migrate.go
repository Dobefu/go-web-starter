package cmd

import (
	"fmt"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	err := database.MigrateUp(getDatabaseConfig())

	if err != nil {
		panic(err)
	}
}

func migrateDown(cmd *cobra.Command, args []string) {
	err := database.MigrateDown(getDatabaseConfig())

	if err != nil {
		panic(err)
	}
}

func migrateVersion(cmd *cobra.Command, args []string) {
	version, err := database.MigrateVersion(getDatabaseConfig())

	if err != nil {
		panic(err)
	}

	fmt.Printf("The migrations are at version %d\n", version)
}

func getDatabaseConfig() database.Config {
	return database.Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
	}
}
