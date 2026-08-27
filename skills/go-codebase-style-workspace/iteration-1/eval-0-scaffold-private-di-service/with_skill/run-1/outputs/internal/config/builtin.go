package config

import "github.com/spf13/viper"

// SetDefaults registers built-in default values for every setting, so the
// service starts with sane values even when the config file omits keys.
// Environment-specific configs (config.dev.yaml, config.prod.yaml, ...)
// override them.
func SetDefaults() {
	viper.SetDefault("logger.level", "info")

	viper.SetDefault("http_server.listen", ":8080")
	viper.SetDefault("http_server.read_timeout", "30s")
	viper.SetDefault("http_server.write_timeout", "30s")
	viper.SetDefault("http_server.read_header_timeout", "10s")
	viper.SetDefault("http_server.idle_timeout", "120s")

	viper.SetDefault("database.dsn", "invoice:invoice@tcp(localhost:3306)/invoice_service?parseTime=true")
	viper.SetDefault("database.max_conn", 25)
	viper.SetDefault("database.idle_conn", 5)
	viper.SetDefault("database.timeout", "30s")

	viper.SetDefault("jwt.secret", "change-me-in-production")

	viper.SetDefault("notifier.webhook_url", "")

	viper.SetDefault("worker.enabled", true)
	viper.SetDefault("worker.jobs_intervals.reconcile_invoices", "5m")

	viper.SetDefault("workers.enabled", false)
	viper.SetDefault("workers.count", 4)
	viper.SetDefault("workers.queue_size", 128)
	viper.SetDefault("workers.timeout", "2m")
	viper.SetDefault("workers.retries", 3)
}
