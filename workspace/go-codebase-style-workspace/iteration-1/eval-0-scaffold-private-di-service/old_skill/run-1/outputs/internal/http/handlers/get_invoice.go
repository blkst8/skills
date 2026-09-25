package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/log"
	"github.com/blkst8/invoice-service/internal/repository"
)

// GetInvoice returns a single invoice by id.
//
// Returns HTTP 404 when the invoice does not exist.
func (h *Handlers) GetInvoice(ctx echo.Context) error {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid invoice id")
	}

	invoice, err := h.svc.Invoice.Get(ctx.Request().Context(), id)
	if err != nil {
		log.Logger.Error("failed to get invoice", zap.Error(err))

		if errors.Is(err, repository.ErrInvoiceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "invoice not found")
		}

		return err
	}

	return ctx.JSON(http.StatusOK, invoice)
}
