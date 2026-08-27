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
	"yourproject/internal/usecase"
)

func UpdateClient(ctx echo.Context) error {
	id64, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	var request usecase.UpdateClientRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	if request.Name == "" || request.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and email are required")
	}

	result, err := app.A.Usecase.Client.Update(ctx.Request().Context(), uint32(id64), request)
	if err != nil {
		log.Logger.Error("failed to update client", zap.Error(err))

		if errors.Is(err, repository.ErrClientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Client not found")
		}

		return err
	}

	return ctx.JSON(http.StatusOK, result)
}
