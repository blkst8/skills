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

// GetClient handles fetching a client by id.
//
// Returns HTTP 200 on success, 400 on invalid id, 404 when the client does
// not exist, 500 on server error.
func (h *Handlers) GetClient(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid client id")
	}

	client, err := h.uc.Client.Get(ctx.Request().Context(), uint32(id))
	if err != nil {
		log.Logger.Error("failed to get client", zap.Error(err))

		if errors.Is(err, repository.ErrClientNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "client not found")
		}

		return err
	}

	return ctx.JSON(http.StatusOK, client)
}
