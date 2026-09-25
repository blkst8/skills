package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/log"
	"github.com/blkst8/invoice-service/internal/repository"
	"github.com/blkst8/invoice-service/internal/usecase"
)

// CreateInvoice handles the creation of a new invoice.
//
// Request body:
//   - customer_id (uint32, required): ID of the billed customer
//   - number (string, required): unique invoice number
//   - amount (float64, required): invoice amount, must be positive
//   - currency (string, optional): ISO currency code, defaults to USD
//   - due_date (time, required): invoice due date
//
// Returns HTTP 200 on success, 400 on validation error, 409 on duplicate
// invoice number, 500 on server error.
func (h *Handlers) CreateInvoice(ctx echo.Context) error {
	var request usecase.CreateInvoiceRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if request.CustomerID == 0 || request.Number == "" || request.Amount <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "customer_id, number and a positive amount are required")
	}

	invoice, err := h.uc.Invoice.Create(ctx.Request().Context(), request)
	if err != nil {
		log.Logger.Error("failed to create invoice", zap.Error(err))

		if errors.Is(err, repository.ErrInvoiceNumberTaken) {
			return echo.NewHTTPError(http.StatusConflict, "invoice number already taken")
		}

		return err
	}

	return ctx.JSON(http.StatusOK, invoice)
}
