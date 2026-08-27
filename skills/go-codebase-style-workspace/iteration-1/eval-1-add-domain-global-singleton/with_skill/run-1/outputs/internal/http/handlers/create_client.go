package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/log"
	"yourproject/internal/usecase"
)

type CreateClientResponse struct {
	ID    uint32 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func CreateClient(ctx echo.Context) error {
	var request usecase.CreateClientRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	if request.Name == "" || request.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and email are required")
	}

	result, err := app.A.Usecase.Client.Create(ctx.Request().Context(), request)
	if err != nil {
		log.Logger.Error("failed to create client", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, CreateClientResponse{
		ID:    result.ID,
		Name:  result.Name,
		Email: result.Email,
	})
}
