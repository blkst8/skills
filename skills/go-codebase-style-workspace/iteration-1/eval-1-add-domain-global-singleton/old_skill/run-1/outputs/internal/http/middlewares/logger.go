// Package middlewares provides reusable Echo HTTP middlewares.
package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ZapLogger logs every request with structured fields, skipping the given
// URLs (e.g. health and metrics endpoints).
func ZapLogger(logger *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
	skip := make(map[string]struct{}, len(skipURLs))
	for _, url := range skipURLs {
		skip[url] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if _, ok := skip[ctx.Path()]; ok {
				return next(ctx)
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
