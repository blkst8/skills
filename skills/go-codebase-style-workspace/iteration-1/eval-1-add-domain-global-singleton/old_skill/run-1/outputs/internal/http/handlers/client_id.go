package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// parseClientID extracts and validates the :id path parameter shared by the
// client handlers.
func parseClientID(ctx echo.Context) (uint32, error) {
	raw := ctx.Param("id")

	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	return uint32(id), nil
}
