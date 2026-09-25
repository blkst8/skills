// Package config loads the YAML application configuration with Viper.
package config

import (
	"time"

	"github.com/spf13/viper"
)

// C is the loaded application configuration.
var C *Config

// Config is the root configuration structure.
type Config struct {
	Logger     Logger     `yaml:"logger"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"database"`
	Telegram   Telegram   `yaml:"telegram"`
}

// Logger configures the zap logger.
type Logger struct {
	Level string `yaml:"level"`
}

// HTTPServer configures the Echo HTTP server.
type HTTPServer struct {
	Listen            string        `yaml:"listen"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

// Database configures the MySQL connection.
type Database struct {
	DSN      string        `yaml:"dsn"`
	MaxConn  int           `yaml:"max_conn"`
	IdleConn int           `yaml:"idle_conn"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Telegram configures the Telegram Bot API client.
type Telegram struct {
	Token string `yaml:"token"`
}

// Load reads the configuration file at configPath into C.
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&C); err != nil {
		return err
	}

	return nil
}
