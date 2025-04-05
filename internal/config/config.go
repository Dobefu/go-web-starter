package config

import (
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/viper"
)

type Server struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type Log struct {
	Level int `mapstructure:"level"`
}

type Site struct {
	Name string `mapstructure:"name"`
}

type Redis struct {
	Enable   bool   `mapstructure:"enable"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	Log      Log      `mapstructure:"log"`
	Site     Site     `mapstructure:"site"`
	Redis    Redis    `mapstructure:"redis"`
}

func GetLogLevel() logger.Level {
	level := viper.GetInt("log.level")

	if level > 0 {
		return logger.Level(level)
	}

	return logger.Level(DefaultConfig.Log.Level)
}

var DefaultConfig = Config{
	Server: Server{
		Port: 4000,
		Host: "localhost",
	},
	Database: Database{
		Host:     "127.0.0.1",
		Port:     2345,
		User:     "root",
		Password: "root",
		DBName:   "db",
	},
	Log: Log{
		Level: int(logger.InfoLevel),
	},
	Site: Site{
		Name: "Go Web Starter",
	},
	Redis: Redis{
		Enable:   true,
		Host:     "127.0.0.1",
		Port:     9736,
		Password: "root",
		DB:       0,
	},
}
