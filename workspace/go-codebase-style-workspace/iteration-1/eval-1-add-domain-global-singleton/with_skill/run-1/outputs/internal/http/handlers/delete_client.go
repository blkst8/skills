package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/log"
	"yourproject/internal/repository"
)

func DeleteClient(ctx echo.Context) error {
	id64, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	err = app.A.Usecase.Client.Delete(ctx.Request().Context(), uint32(id64))
	if err != nil {
		log.Logger.Error("failed to delete client", zap.Error(err))

		if errors.Is(err, repository.ErrClientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Client not found")
		}

		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}
