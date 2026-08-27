// Package log provides the application-wide structured logger.
//
// Application code must use log.Logger with zap fields
// (zap.Error(err), zap.String("k", v), ...) — never fmt.Println or the
// standard library log package.
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	var err error
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	Logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

// SetLevel adjusts the logger level after config load; invalid levels fall
// back to info.
func SetLevel(level string) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}
	Logger = Logger.WithOptions(zap.IncreaseLevel(lvl))
}
