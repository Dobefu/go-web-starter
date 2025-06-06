package config

import (
	"encoding/base64"

	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

var defaultHost = "127.0.0.1"

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

type Email struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Identity string `mapstructure:"identity"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type Log struct {
	Level int `mapstructure:"level"`
}

type Site struct {
	Name  string `mapstructure:"name"`
	Host  string `mapstructure:"host"`
	Email string `mapstructure:"email"`
}

type Redis struct {
	Enable   bool   `mapstructure:"enable"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Session struct {
	Secret string `mapstructure:"secret"`
}

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	Email    Email    `mapstructure:"email"`
	Log      Log      `mapstructure:"log"`
	Site     Site     `mapstructure:"site"`
	Redis    Redis    `mapstructure:"redis"`
	Session  Session  `mapstructure:"session"`
}

func GetLogLevel() logger.Level {
	level := viper.GetInt("log.level")

	if viper.IsSet("log.level") {
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
		Host:     defaultHost,
		Port:     2345,
		User:     "root",
		Password: "root",
		DBName:   "db",
	},
	Email: Email{
		Host:     defaultHost,
		Port:     5201,
		Identity: "",
		User:     "",
		Password: "",
	},
	Log: Log{
		Level: int(logger.InfoLevel),
	},
	Site: Site{
		Name:  "Go Web Starter",
		Host:  "http://localhost:4000",
		Email: "info@example.com",
	},
	Redis: Redis{
		Enable:   true,
		Host:     defaultHost,
		Port:     9736,
		Password: "root",
		DB:       0,
	},
	Session: Session{
		Secret: base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)),
	},
}
