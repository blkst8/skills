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

// GetClient handles fetching a single client by ID.
//
// Returns HTTP 200 on success, 400 on invalid ID, 404 when the client does
// not exist, 500 on server error.
func GetClient(ctx echo.Context) error {
	id, err := parseClientID(ctx)
	if err != nil {
		return err
	}

	client, err := app.A.Service.Client.Get(ctx.Request().Context(), id)
	if err != nil {
		log.Logger.Error("failed to get client", zap.Error(err))

		if errors.Is(err, repository.ErrClientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "client not found")
		}

		return err
	}

	return ctx.JSON(http.StatusOK, client)
}
