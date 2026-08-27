package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Healthz reports service liveness.
func Healthz(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Metrics exposes Prometheus metrics.
func Metrics(ctx echo.Context) error {
	promhttp.Handler().ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
