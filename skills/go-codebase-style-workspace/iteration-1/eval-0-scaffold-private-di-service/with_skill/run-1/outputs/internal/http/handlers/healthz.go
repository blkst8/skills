package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Healthz reports service liveness.
func Healthz(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
