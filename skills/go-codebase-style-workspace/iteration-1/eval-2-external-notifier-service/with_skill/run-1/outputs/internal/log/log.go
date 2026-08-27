// Package log provides the shared zap logger for the application.
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the application-wide structured logger.
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
