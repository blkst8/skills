package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/log"
	"yourproject/internal/models"
	"yourproject/internal/repository"
	"yourproject/internal/usecase"
)

// UpdateClientRequest is the request body for updating a client.
type UpdateClientRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateClient handles updating an existing client.
//
// It binds the request, updates the stored client through the client usecase
// and returns the fresh record.
//
// Returns HTTP 200 on success, 400 on invalid ID or validation error,
// 404 when the client does not exist, 409 when the new email already exists,
// 500 on server error.
func UpdateClient(ctx echo.Context) error {
	id, err := parseClientID(ctx)
	if err != nil {
		return err
	}

	var request UpdateClientRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	client := models.Client{
		ID:    id,
		Name:  request.Name,
		Email: request.Email,
	}

	if err := app.A.Service.Client.Update(ctx.Request().Context(), client); err != nil {
		log.Logger.Error("failed to update client", zap.Error(err))

		switch {
		case errors.Is(err, usecase.ErrInvalidClientInput):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrClientNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "client not found")
		case errors.Is(err, repository.ErrClientAlreadyExists):
			return echo.NewHTTPError(http.StatusConflict, "client already exists")
		default:
			return err
		}
	}

	updated, err := app.A.Service.Client.Get(ctx.Request().Context(), id)
	if err != nil {
		log.Logger.Error("failed to get updated client", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, updated)
}
