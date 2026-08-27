// Package middlewares contains the Echo middlewares used by the HTTP server.
package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ZapLogger returns a middleware that logs every request with structured
// fields. Paths listed in skipURLs are not logged.
func ZapLogger(logger *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
	skip := make(map[string]struct{}, len(skipURLs))
	for _, u := range skipURLs {
		skip[u] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			start := time.Now()

			err := next(ctx)

			if _, ok := skip[ctx.Path()]; !ok {
				logger.Info("request",
					zap.String("method", ctx.Request().Method),
					zap.String("uri", ctx.Request().RequestURI),
					zap.Int("status", ctx.Response().Status),
					zap.Duration("latency", time.Since(start)),
				)
			}

			return err
		}
	}
}
