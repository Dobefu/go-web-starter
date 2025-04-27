package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clearCacheCmd = &cobra.Command{
	Use:   "clear-cache",
	Short: "Clear the Redis cache",
	Long:  `Clear all keys from the Redis cache database.`,
	Run:   runClearCacheCmd,
}

func init() {
	rootCmd.AddCommand(clearCacheCmd)
}

func runClearCacheCmd(cmd *cobra.Command, args []string) {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	if !viper.GetBool("redis.enable") {
		log.Error("Redis is not enabled in configuration", nil)
		return
	}

	redisConfig := config.Redis{
		Enable:   viper.GetBool("redis.enable"),
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}

	log.Debug("Initializing Redis connection", logger.Fields{
		"host": redisConfig.Host,
		"port": redisConfig.Port,
		"db":   redisConfig.DB,
	})

	redisClient, err := redis.New(redisConfig, log)

	if err != nil {
		log.Error("Failed to initialize Redis", logger.Fields{
			"error": err.Error(),
		})

		return
	}

	defer func() { _ = redisClient.Close() }()

	ctx := context.Background()
	log.Debug("Clearing Redis cache", nil)
	result, err := redisClient.FlushDB(ctx)

	if err != nil {
		log.Error("Failed to clear Redis cache", logger.Fields{
			"error": err.Error(),
		})

		return
	}

	log.Debug("Redis cache cleared successfully", logger.Fields{
		"result": result.Val(),
	})

	fmt.Println("Redis cache cleared successfully")
}
