// Package log provides the global zap logger of the application.
//
// Always use log.Logger instead of fmt.Println or the standard log package
// in application code.
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger instance.
var Logger *zap.Logger

func init() {
	var err error
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}
