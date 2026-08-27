// Package middlewares provides Echo HTTP middleware.
package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ZapLogger logs every request with structured fields. Requests to skipURLs
// are passed through unlogged.
func ZapLogger(logger *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			for _, skip := range skipURLs {
				if ctx.Request().URL.Path == skip {
					return next(ctx)
				}
			}

			start := time.Now()
			err := next(ctx)

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
