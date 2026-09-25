package handlers

import (
	"github.com/labstack/echo/v4"

	"github.com/blkst8/invoice-service/internal/metrics"
)

// Metrics exposes Prometheus metrics.
func Metrics(ctx echo.Context) error {
	metrics.Handler().ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
