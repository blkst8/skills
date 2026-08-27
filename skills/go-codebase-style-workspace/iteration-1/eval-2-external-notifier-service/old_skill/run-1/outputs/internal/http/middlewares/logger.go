// Package middlewares provides the Echo middlewares used by the HTTP server.
package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ZapLogger logs every request with structured fields. Paths listed in
// skipURLs (for example /healthz and /metrics) are not logged.
func ZapLogger(logger *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			start := time.Now()

			err := next(ctx)

			if isSkipped(ctx.Request().URL.Path, skipURLs) {
				return err
			}

			logger.Info("request",
				zap.String("method", ctx.Request().Method),
				zap.String("uri", ctx.Request().RequestURI),
				zap.Int("status", ctx.Response().Status),
				zap.Duration("latency", time.Since(start)),
			)

			return err
		}
	}
}

// isSkipped reports whether path matches one of the skipURLs.
func isSkipped(path string, skipURLs []string) bool {
	for _, skipURL := range skipURLs {
		if path == skipURL {
			return true
		}
	}

	return false
}
