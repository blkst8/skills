// Package config loads the application configuration from a YAML file using
// viper, with built-in defaults for every setting.
package config

import (
	"time"

	"github.com/spf13/viper"
)

// C is the loaded application configuration, set by Load.
var C *Config

type Config struct {
	Logger     Logger     `yaml:"logger" mapstructure:"logger"`
	HTTPServer HTTPServer `yaml:"http_server" mapstructure:"http_server"`
	Database   Database   `yaml:"database" mapstructure:"database"`
	JWT        JWT        `yaml:"jwt" mapstructure:"jwt"`
	Notifier   Notifier   `yaml:"notifier" mapstructure:"notifier"`
	Worker     Worker     `yaml:"worker" mapstructure:"worker"`
	Workers    Workers    `yaml:"workers" mapstructure:"workers"`
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
}

type Database struct {
	DSN      string        `yaml:"dsn" mapstructure:"dsn"`
	MaxConn  int           `yaml:"max_conn" mapstructure:"max_conn"`
	IdleConn int           `yaml:"idle_conn" mapstructure:"idle_conn"`
	Timeout  time.Duration `yaml:"timeout" mapstructure:"timeout"`
}

type JWT struct {
	Secret string `yaml:"secret" mapstructure:"secret"`
}

type Notifier struct {
	WebhookURL string `yaml:"webhook_url" mapstructure:"webhook_url"`
}

// Worker configures the interval-based background jobs (internal/worker).
type Worker struct {
	Enabled       bool          `yaml:"enabled" mapstructure:"enabled"`
	JobsIntervals JobsIntervals `yaml:"jobs_intervals" mapstructure:"jobs_intervals"`
}

type JobsIntervals struct {
	ReconcileInvoices time.Duration `yaml:"reconcile_invoices" mapstructure:"reconcile_invoices"`
}

// Workers configures the worker pool (internal/workers).
type Workers struct {
	Enabled   bool          `yaml:"enabled" mapstructure:"enabled"`
	Count     int           `yaml:"count" mapstructure:"count"`
	QueueSize int           `yaml:"queue_size" mapstructure:"queue_size"`
	Timeout   time.Duration `yaml:"timeout" mapstructure:"timeout"`
	Retries   int           `yaml:"retries" mapstructure:"retries"`
}

// Load reads the YAML config at configPath into C, on top of the built-in
// defaults registered by SetDefaults.
func Load(configPath string) error {
	SetDefaults()

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
