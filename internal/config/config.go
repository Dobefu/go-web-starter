package config

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
}

var DefaultConfig = Config{
	Server: struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	}{
		Port: 8080,
		Host: "localhost",
	},
	Database: struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
	}{
		Host:   "localhost",
		Port:   5432,
		User:   "postgres",
		DBName: "app",
	},
}
