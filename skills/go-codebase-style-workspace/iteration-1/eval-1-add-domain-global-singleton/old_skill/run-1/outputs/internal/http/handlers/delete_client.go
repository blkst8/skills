package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/log"
	"yourproject/internal/repository"
)

// DeleteClient handles deleting a client by ID.
//
// Returns HTTP 204 on success, 400 on invalid ID, 404 when the client does
// not exist, 500 on server error.
func DeleteClient(ctx echo.Context) error {
	id, err := parseClientID(ctx)
	if err != nil {
		return err
	}

	if err := app.A.Service.Client.Delete(ctx.Request().Context(), id); err != nil {
		log.Logger.Error("failed to delete client", zap.Error(err))

		if errors.Is(err, repository.ErrClientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "client not found")
		}

		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}
