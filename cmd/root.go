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
	log     *logger.Logger
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
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "Enable verbose output (-v info, -vv debug, -vvv trace)")
}

func createDefaultConfigFileIfNotExist(filePath string) error {
	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		return err
	}

	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	configFileContent, err := toml.Marshal(config.DefaultConfig)

	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(filePath, configFileContent, 0666); err != nil {
		return fmt.Errorf("failed to write default config file %s: %w", filePath, err)
	}

	fmt.Printf("Default configuration file created at %s\n", filePath)
	return nil
}

func initConfig() {
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("toml")
		viper.SetConfigName(defaultConfigFileName)

		err := createDefaultConfigFileIfNotExist(defaultConfigFileName)
		if err != nil && !os.IsNotExist(err) {
			panic(fmt.Errorf("error ensuring default config file exists: %w", err))
		}
	}

	err := viper.ReadInConfig()

	if err != nil {
		_, isConfigFileNotFound := err.(viper.ConfigFileNotFoundError)

		if !isConfigFileNotFound {
			tmpLog := logger.New(logger.WarnLevel, os.Stderr)
			tmpLog.Error("Error reading config file", map[string]any{"file": viper.ConfigFileUsed(), "error": err.Error()})
		} else if cfgFile != "" {
			tmpLog := logger.New(logger.ErrorLevel, os.Stderr)
			tmpLog.Error("Specified config file not found", map[string]any{"file": cfgFile, "error": err.Error()})
			os.Exit(1)
		}
	}

	logLevel := config.GetLogLevel()

	if verbose == 1 {
		logLevel = logger.InfoLevel
	} else if verbose == 2 {
		logLevel = logger.DebugLevel
	} else if verbose >= 3 {
		logLevel = logger.TraceLevel
	}

	if quiet {
		logLevel = logger.ErrorLevel
	}

	viper.Set("log.level", int(logLevel))

	log = logger.New(logLevel, os.Stdout)
	log.Trace("Using configuration file", map[string]any{"file": viper.ConfigFileUsed()})
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
