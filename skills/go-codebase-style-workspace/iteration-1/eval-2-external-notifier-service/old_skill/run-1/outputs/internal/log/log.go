// Package log provides the application-wide zap logger.
//
// Application code must always use log.Logger with structured fields instead
// of the standard library log package or fmt.Println.
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the process-wide logger, initialized by init.
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
