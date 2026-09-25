package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/log"
)

// ListInvoices returns a paginated list of invoices, newest first.
//
// Query parameters:
//   - limit (int, optional): page size, clamped by the service
//   - offset (int, optional): page offset
func (h *Handlers) ListInvoices(ctx echo.Context) error {
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 0
	}

	offset, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil {
		offset = 0
	}

	invoices, err := h.svc.Invoice.List(ctx.Request().Context(), limit, offset)
	if err != nil {
		log.Logger.Error("failed to list invoices", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, invoices)
}
