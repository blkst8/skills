package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/log"
	"github.com/blkst8/client-service/internal/usecase"
)

// CreateClientRequest is the body of the create-client request.
type CreateClientRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	TelegramID string `json:"telegram_id"`
}

// CreateClientResponse is the body of the create-client response.
type CreateClientResponse struct {
	ID         uint32 `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	TelegramID string `json:"telegram_id"`
}

// CreateClient handles the creation of a new client.
//
// It validates the request, delegates to the client usecase (which persists
// the client and best-effort sends a Telegram welcome message), and returns
// the created client.
//
// Returns HTTP 200 on success, 400 on validation error, 500 on server error.
func (h *Handlers) CreateClient(ctx echo.Context) error {
	var request CreateClientRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	if request.Name == "" || request.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and email are required")
	}

	client, err := h.uc.Client.Create(ctx.Request().Context(), usecase.CreateClientRequest{
		Name:       request.Name,
		Email:      request.Email,
		TelegramID: request.TelegramID,
	})
	if err != nil {
		log.Logger.Error("failed to create client", zap.Error(err))

		return err
	}

	return ctx.JSON(http.StatusOK, CreateClientResponse{
		ID:         client.ID,
		Name:       client.Name,
		Email:      client.Email,
		TelegramID: client.TelegramID,
	})
}
