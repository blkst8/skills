package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics exposes Prometheus metrics.
func Metrics(ctx echo.Context) error {
	promhttp.Handler().ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
