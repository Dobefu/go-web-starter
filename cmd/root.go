package cmd

import (
	"os"
	"path/filepath"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "./app",
	Short: "The main command to manage the website",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "The config file to use (default: ./config.toml)")
}

func initConfig() {
	defaultConfigFile := "config.toml"
	viper.AddConfigPath(".")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("toml")
		viper.SetConfigName(defaultConfigFile)
	}

	viper.AutomaticEnv()

	if cfgFile == "" {
		if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
			viper.Set("database.host", config.DefaultConfig.Database.Host)
			viper.Set("database.port", config.DefaultConfig.Database.Port)
			viper.Set("database.user", config.DefaultConfig.Database.User)
			viper.Set("database.password", config.DefaultConfig.Database.Password)
			viper.Set("database.dbname", config.DefaultConfig.Database.DBName)
			viper.Set("server.port", config.DefaultConfig.Server.Port)
			viper.Set("server.host", config.DefaultConfig.Server.Host)

			dir := filepath.Dir(defaultConfigFile)

			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}

			if err := viper.WriteConfigAs("config.toml"); err != nil {
				panic(err)
			}
		}
	}

	_ = viper.ReadInConfig()
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
