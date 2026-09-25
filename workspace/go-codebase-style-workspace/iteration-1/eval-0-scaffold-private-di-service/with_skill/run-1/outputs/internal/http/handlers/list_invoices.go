package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/log"
)

// ListInvoices returns all invoices.
//
// Returns HTTP 200 on success, 500 on server error.
func (h *Handlers) ListInvoices(ctx echo.Context) error {
	invoices, err := h.uc.Invoice.List(ctx.Request().Context())
	if err != nil {
		log.Logger.Error("failed to list invoices", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, invoices)
}
