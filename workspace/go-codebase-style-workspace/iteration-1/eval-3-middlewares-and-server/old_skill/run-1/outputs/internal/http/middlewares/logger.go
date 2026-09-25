package middlewares

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ZapLogger returns an Echo middleware that logs every request with
// structured zap fields: method, URI, response status, and latency.
//
// Requests whose URL path matches one of skipURLs are passed through
// without being logged — use it for high-frequency endpoints such as
// health checks and metrics scraping.
func ZapLogger(log *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			for _, skipURL := range skipURLs {
				if ctx.Request().URL.Path == skipURL {
					return next(ctx)
				}
			}

			start := time.Now()
			err := next(ctx)

			// Echo writes error responses only after the middleware chain
			// unwinds, so the status must be derived from the returned
			// error — otherwise every failing request is logged as 200.
			status := ctx.Response().Status
			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok {
					status = httpErr.Code
				} else {
					status = http.StatusInternalServerError
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
