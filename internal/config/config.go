package config

import (
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
	} `mapstructure:"database"`

	Log struct {
		Level int `mapstructure:"level"`
	} `mapstructure:"log"`
}

func GetLogLevel() logger.Level {
	level := viper.GetInt("log.level")

	if level > 0 {
		return logger.Level(level)
	}

	return logger.Level(DefaultConfig.Log.Level)
}

var DefaultConfig = Config{
	Server: struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	}{
		Port: 4000,
		Host: "localhost",
	},
	Database: struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
	}{
		Host:     "127.0.0.1",
		Port:     2345,
		User:     "root",
		Password: "root",
		DBName:   "db",
	},
	Log: struct {
		Level int `mapstructure:"level"`
	}{
		Level: int(logger.InfoLevel),
	},
}
