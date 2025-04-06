package cmd

import (
	"os"
	"path/filepath"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
		_, err := os.Stat(defaultConfigFile)

		if os.IsNotExist(err) {
			viper.Set("database.host", config.DefaultConfig.Database.Host)
			viper.Set("database.port", config.DefaultConfig.Database.Port)
			viper.Set("database.user", config.DefaultConfig.Database.User)
			viper.Set("database.password", config.DefaultConfig.Database.Password)
			viper.Set("database.dbname", config.DefaultConfig.Database.DBName)
			viper.Set("server.port", config.DefaultConfig.Server.Port)
			viper.Set("server.host", config.DefaultConfig.Server.Host)
			viper.Set("log.level", config.DefaultConfig.Log.Level)
			viper.Set("site.name", config.DefaultConfig.Site.Name)
			viper.Set("redis.enable", config.DefaultConfig.Redis.Enable)
			viper.Set("redis.host", config.DefaultConfig.Redis.Host)
			viper.Set("redis.port", config.DefaultConfig.Redis.Port)
			viper.Set("redis.password", config.DefaultConfig.Redis.Password)
			viper.Set("redis.db", config.DefaultConfig.Redis.DB)

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

	if verbose > 0 {
		level := logger.DebugLevel

		if verbose >= 3 {
			level = logger.TraceLevel
		}

		viper.Set("log.level", level)
	}
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
