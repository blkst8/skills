// Package config loads application configuration from a YAML file with Viper.
//
// The loaded configuration is stored in the C singleton. Real secrets (such as
// the Telegram bot token) must never be committed; provide them through an
// environment-specific config file and keep only config.example.yaml in git.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// C is the loaded application configuration. It is set by Load.
var C *Config

// Config is the root configuration structure.
type Config struct {
	Logger     Logger     `yaml:"logger"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"database"`
	Telegram   Telegram   `yaml:"telegram"`
}

// Logger holds the logging settings.
type Logger struct {
	Level string `yaml:"level"`
}

// HTTPServer holds the HTTP server settings.
type HTTPServer struct {
	Listen            string        `yaml:"listen"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

// Database holds the database connection settings.
type Database struct {
	DSN      string        `yaml:"dsn"`
	MaxConn  int           `yaml:"max_conn"`
	IdleConn int           `yaml:"idle_conn"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Telegram holds the settings of the Telegram Bot API notifier client.
type Telegram struct {
	// Enabled turns Telegram notifications on or off. When disabled, a
	// no-op notifier is wired so client creation is never affected.
	Enabled bool `yaml:"enabled"`
	// BotToken is the token issued by @BotFather.
	BotToken string `yaml:"bot_token"`
	// APIBaseURL is the root URL of the Bot API. Override it in tests or
	// for self-hosted Bot API servers.
	APIBaseURL string `yaml:"api_base_url"`
	// Timeout bounds every call to the Telegram Bot API so a Telegram
	// outage can never hang client creation.
	Timeout time.Duration `yaml:"timeout"`
}

// Load reads the YAML configuration file at configPath and unmarshals it into C.
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	C = &Config{}
	if err := viper.Unmarshal(C); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return nil
}
