package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics serves the Prometheus metrics endpoint.
func Metrics(ctx echo.Context) error {
	promhttp.Handler().ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
