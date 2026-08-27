package middlewares

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ZapLogger returns a middleware that writes one structured zap entry per
// request. Paths listed in skipURLs (e.g. /healthz, /metrics) are served
// without being logged.
func ZapLogger(log *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
	skip := make(map[string]struct{}, len(skipURLs))
	for _, url := range skipURLs {
		skip[url] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if _, ok := skip[ctx.Request().URL.Path]; ok {
				return next(ctx)
			}

			start := time.Now()
			err := next(ctx)

			status := ctx.Response().Status
			if err != nil {
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					status = httpErr.Code
				}
			}

			log.Info("request",
				zap.String("method", ctx.Request().Method),
				zap.String("uri", ctx.Request().RequestURI),
				zap.Int("status", status),
				zap.Duration("latency", time.Since(start)),
			)

			return err
		}
	}
}
