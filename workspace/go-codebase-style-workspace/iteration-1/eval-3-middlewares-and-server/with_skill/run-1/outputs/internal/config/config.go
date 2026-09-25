package config

import (
	"time"

	"github.com/spf13/viper"
)

// C holds the loaded application configuration.
var C *Config

type Config struct {
	Logger     Logger     `yaml:"logger"`
	HTTPServer HTTPServer `yaml:"http_server"`
	JWT        JWT        `yaml:"jwt"`
	Database   Database   `yaml:"database"`
	Worker     Worker     `yaml:"worker"`
	Workers    Workers    `yaml:"workers"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type HTTPServer struct {
	Listen            string        `yaml:"listen"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

// JWT holds the token-signing settings consumed by the authentication
// middleware (internal/http/middlewares).
type JWT struct {
	Secret string `yaml:"secret"`
}

type Database struct {
	DSN      string        `yaml:"dsn"`
	MaxConn  int           `yaml:"max_conn"`
	IdleConn int           `yaml:"idle_conn"`
	Timeout  time.Duration `yaml:"timeout"`
}

type Worker struct {
	Enabled       bool          `yaml:"enabled"`
	JobsIntervals JobsIntervals `yaml:"jobs_intervals"`
}

type JobsIntervals struct {
	SyncDatabases time.Duration `yaml:"sync_databases"`
	CleanupJobs   time.Duration `yaml:"cleanup_jobs"`
}

// Workers configures the worker pool (internal/workers).
type Workers struct {
	Enabled   bool          `yaml:"enabled"`
	Count     int           `yaml:"count"`
	QueueSize int           `yaml:"queue_size"`
	Timeout   time.Duration `yaml:"timeout"`
	Retries   int           `yaml:"retries"`
}

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
