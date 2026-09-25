// Package config loads the application configuration with Viper.
//
// C is the global configuration singleton, populated by Load.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// C is the global configuration instance.
var C *Config

// Config is the root configuration structure.
type Config struct {
	Logger     Logger     `yaml:"logger"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"database"`
	Worker     Worker     `yaml:"worker"`
}

// Logger holds the logging configuration.
type Logger struct {
	Level string `yaml:"level"`
}

// HTTPServer holds the HTTP server configuration.
type HTTPServer struct {
	Listen            string        `yaml:"listen"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

// Database holds the database connection configuration.
type Database struct {
	DSN      string        `yaml:"dsn"`
	MaxConn  int           `yaml:"max_conn"`
	IdleConn int           `yaml:"idle_conn"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Worker holds the background worker configuration.
type Worker struct {
	Enabled       bool          `yaml:"enabled"`
	JobsIntervals JobsIntervals `yaml:"jobs_intervals"`
}

// JobsIntervals holds the interval of each background job.
type JobsIntervals struct {
	SyncDatabases time.Duration `yaml:"sync_databases"`
	CleanupJobs   time.Duration `yaml:"cleanup_jobs"`
}

// Load reads the YAML configuration file and populates the global C.
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&C); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
