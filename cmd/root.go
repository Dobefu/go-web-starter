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
		cmd.Usage()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "The config file to use (default: ./config.toml)")
}

func initConfig() {
	viper.AddConfigPath(".")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("toml")
		viper.SetConfigName("config.toml")
	}

	viper.AutomaticEnv()

	if cfgFile == "" {
		if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
			viper.Set("server.port", config.DefaultConfig.Server.Port)
			viper.Set("server.host", config.DefaultConfig.Server.Host)

			dir := filepath.Dir("config.toml")

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
