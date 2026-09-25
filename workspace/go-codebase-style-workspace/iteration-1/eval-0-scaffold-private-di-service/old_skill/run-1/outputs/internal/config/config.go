// Package config loads the application configuration from a YAML file.
//
// The loaded configuration is exposed through the package-level C variable.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// C holds the application configuration after Load succeeds.
var C *Config

type Config struct {
	Logger     Logger     `yaml:"logger" mapstructure:"logger"`
	HTTPServer HTTPServer `yaml:"http_server" mapstructure:"http_server"`
	Database   Database   `yaml:"database" mapstructure:"database"`
	Worker     Worker     `yaml:"worker" mapstructure:"worker"`
}

type Logger struct {
	Level string `yaml:"level" mapstructure:"level"`
}

type HTTPServer struct {
	Listen            string        `yaml:"listen" mapstructure:"listen"`
	ReadTimeout       time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" mapstructure:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	JWTSecret         string        `yaml:"jwt_secret" mapstructure:"jwt_secret"`
}

type Database struct {
	DSN      string        `yaml:"dsn" mapstructure:"dsn"`
	MaxConn  int           `yaml:"max_conn" mapstructure:"max_conn"`
	IdleConn int           `yaml:"idle_conn" mapstructure:"idle_conn"`
	Timeout  time.Duration `yaml:"timeout" mapstructure:"timeout"`
}

type Worker struct {
	Enabled       bool          `yaml:"enabled" mapstructure:"enabled"`
	JobsIntervals JobsIntervals `yaml:"jobs_intervals" mapstructure:"jobs_intervals"`
}

type JobsIntervals struct {
	ReconcileInvoices time.Duration `yaml:"reconcile_invoices" mapstructure:"reconcile_invoices"`
}

// Load reads the YAML configuration file, unmarshals it into C and applies
// defaults for any required values left empty.
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	cfg.applyDefaults()

	C = cfg
	return nil
}

// MigrateDSN returns the database DSN in the URL form expected by golang-migrate.
func (d Database) MigrateDSN() string {
	return "mysql://" + d.DSN
}
