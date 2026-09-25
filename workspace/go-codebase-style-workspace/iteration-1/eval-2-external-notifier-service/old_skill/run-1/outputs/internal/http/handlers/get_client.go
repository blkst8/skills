package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/log"
	"github.com/blkst8/client-service/internal/repository"
)

// GetClientResponse is the body of the get-client response.
type GetClientResponse struct {
	ID         uint32 `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	TelegramID string `json:"telegram_id"`
}

// GetClient returns a client by ID.
//
// It maps the repository sentinel error ErrClientNotFound to HTTP 404 and
// every other failure to HTTP 500.
func (h *Handlers) GetClient(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	client, err := h.svc.Client.Get(ctx.Request().Context(), uint32(id))
	if err != nil {
		log.Logger.Error("failed to get client", zap.Error(err), zap.Uint64("id", id))

		if errors.Is(err, repository.ErrClientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "client not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get client")
	}

	return ctx.JSON(http.StatusOK, GetClientResponse{
		ID:         client.ID,
		Name:       client.Name,
		Email:      client.Email,
		TelegramID: client.TelegramID,
	})
}
