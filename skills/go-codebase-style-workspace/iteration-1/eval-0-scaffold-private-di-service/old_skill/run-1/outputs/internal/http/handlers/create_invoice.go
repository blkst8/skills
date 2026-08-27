package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/log"
	"github.com/blkst8/invoice-service/internal/services"
)

type CreateInvoiceRequest struct {
	CustomerID  int64     `json:"customer_id"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	DueAt       time.Time `json:"due_at"`
}

type CreateInvoiceResponse struct {
	ID     int64  `json:"id"`
	Number string `json:"number"`
}

// CreateInvoice handles the creation of a new invoice.
//
// Request body:
//   - customer_id (int64, required): id of the customer to bill
//   - amount_cents (int64, required): invoice amount in minor currency units
//   - currency (string, optional): ISO 4217 code, defaults to USD
//   - due_at (time, required): due date of the invoice
//
// Returns HTTP 201 on success, 400 on validation error, 500 on server error.
func (h *Handlers) CreateInvoice(ctx echo.Context) error {
	var req CreateInvoiceRequest
	if err := ctx.Bind(&req); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	invoice, err := h.svc.Invoice.Create(ctx.Request().Context(), services.CreateInvoiceRequest{
		CustomerID:  req.CustomerID,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		DueAt:       req.DueAt,
	})
	if err != nil {
		log.Logger.Error("failed to create invoice", zap.Error(err))

		if errors.Is(err, services.ErrInvalidInvoice) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid invoice request")
		}

		return err
	}

	return ctx.JSON(http.StatusCreated, CreateInvoiceResponse{ID: invoice.ID, Number: invoice.Number})
}
