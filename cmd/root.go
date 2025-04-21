package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultConfigFileName = "config.toml"

var (
	cfgFile string
	quiet   bool
	verbose int
)

var rootCmd = &cobra.Command{
	Use:   "./app",
	Short: "Manage the website",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "The config file to use (default: ./config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "Enable verbose output (use -vv for debug output, -vvv for trace output)")
}

func initConfig() {
	defaultConfigFile := defaultConfigFileName
	viper.AddConfigPath(".")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("toml")
		viper.SetConfigName(defaultConfigFile)
	}

	viper.AutomaticEnv()

	if cfgFile == "" {
		_, err := os.Stat(defaultConfigFile)

		if os.IsNotExist(err) {
			dir := filepath.Dir(defaultConfigFile)

			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}

			configFileContent, err := toml.Marshal(config.DefaultConfig)

			if err != nil {
				panic(fmt.Errorf("failed to marshal default config: %w", err))
			}

			if err := os.WriteFile(defaultConfigFile, configFileContent, 0666); err != nil {
				panic(fmt.Errorf("failed to write default config file: %w", err))
			}
		}
	}

	err := viper.ReadInConfig()

	if err != nil {
		if !os.IsNotExist(err) {
			tmpLog := logger.New(logger.WarnLevel, os.Stderr)
			tmpLog.Error("Error reading config file", map[string]any{"error": err.Error()})
		}
	}

	if verbose > 0 {
		level := logger.DebugLevel

		if verbose >= 3 {
			level = logger.TraceLevel
		}

		viper.Set("log.level", level)
	}

	log := logger.New(config.GetLogLevel(), os.Stdout)
	log.Trace("Starting the application", nil)
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
