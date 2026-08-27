package config

import "time"

// Default returns a Config populated with sane local-development defaults.
func Default() *Config {
	return &Config{
		Logger: Logger{
			Level: "info",
		},
		HTTPServer: HTTPServer{
			Listen:            ":8080",
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		Database: Database{
			DSN:      "invoice:invoice@tcp(localhost:3306)/invoices?parseTime=true&multiStatements=true",
			MaxConn:  25,
			IdleConn: 5,
			Timeout:  30 * time.Second,
		},
		Worker: Worker{
			Enabled: true,
			JobsIntervals: JobsIntervals{
				ReconcileInvoices: 5 * time.Minute,
			},
		},
	}
}

// applyDefaults fills in the defaults for any required field left empty by
// the configuration file.
func (c *Config) applyDefaults() {
	d := Default()

	if c.Logger.Level == "" {
		c.Logger.Level = d.Logger.Level
	}
	if c.HTTPServer.Listen == "" {
		c.HTTPServer.Listen = d.HTTPServer.Listen
	}
	if c.HTTPServer.ReadTimeout == 0 {
		c.HTTPServer.ReadTimeout = d.HTTPServer.ReadTimeout
	}
	if c.HTTPServer.WriteTimeout == 0 {
		c.HTTPServer.WriteTimeout = d.HTTPServer.WriteTimeout
	}
	if c.HTTPServer.ReadHeaderTimeout == 0 {
		c.HTTPServer.ReadHeaderTimeout = d.HTTPServer.ReadHeaderTimeout
	}
	if c.HTTPServer.IdleTimeout == 0 {
		c.HTTPServer.IdleTimeout = d.HTTPServer.IdleTimeout
	}
	if c.Database.DSN == "" {
		c.Database.DSN = d.Database.DSN
	}
	if c.Database.MaxConn == 0 {
		c.Database.MaxConn = d.Database.MaxConn
	}
	if c.Database.IdleConn == 0 {
		c.Database.IdleConn = d.Database.IdleConn
	}
	if c.Database.Timeout == 0 {
		c.Database.Timeout = d.Database.Timeout
	}
	if c.Worker.JobsIntervals.ReconcileInvoices == 0 {
		c.Worker.JobsIntervals.ReconcileInvoices = d.Worker.JobsIntervals.ReconcileInvoices
	}
}
