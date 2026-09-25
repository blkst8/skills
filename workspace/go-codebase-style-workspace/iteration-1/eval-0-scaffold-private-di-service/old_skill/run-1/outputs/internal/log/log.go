// Package log provides the shared zap logger for the application.
//
// Application code must log through log.Logger with structured fields;
// never use fmt.Println or the standard library log package.
package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the application-wide structured logger.
var Logger *zap.Logger

func init() {
	var err error
	Logger, err = newLogger(zapcore.InfoLevel)
	if err != nil {
		panic(err)
	}
}

// SetLevel rebuilds Logger at the given level ("debug", "info", "warn", ...).
func SetLevel(level string) error {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	logger, err := newLogger(lvl)
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	Logger = logger
	return nil
}

func newLogger(level zapcore.Level) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build()
}
