package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/log"
	"github.com/blkst8/client-service/internal/services"
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
// It validates the request, stores the client, and triggers the Telegram
// welcome message inside the service. Notification problems never fail this
// request: the service logs them and returns the created client.
//
// Request body:
//   - name (string, required): client name
//   - email (string, required): client email
//   - telegram_id (string, optional): Telegram chat ID for the welcome message
//
// Returns HTTP 201 on success, 400 on validation error, 500 on server error.
func (h *Handlers) CreateClient(ctx echo.Context) error {
	var request CreateClientRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	if request.Name == "" || request.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and email are required")
	}

	client, err := h.svc.Client.Create(ctx.Request().Context(), services.CreateClientRequest{
		Name:       request.Name,
		Email:      request.Email,
		TelegramID: request.TelegramID,
	})
	if err != nil {
		log.Logger.Error("failed to create client", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create client")
	}

	return ctx.JSON(http.StatusCreated, CreateClientResponse{
		ID:         client.ID,
		Name:       client.Name,
		Email:      client.Email,
		TelegramID: client.TelegramID,
	})
}
