// Package handlers contains the background job handlers.
//
// With the private dependency injection pattern every job handler is a
// struct receiving the service bundle; each job lives in its own file.
package handlers

import (
	"context"

	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/app"
	"github.com/blkst8/invoice-service/internal/log"
)

type reconcileInvoices struct {
	svc *app.Service
}

// NewReconcileInvoices returns the invoice reconciliation job handler.
func NewReconcileInvoices(svc *app.Service) *reconcileInvoices {
	return &reconcileInvoices{svc: svc}
}

// Handle marks overdue invoices; it runs on the configured interval.
func (h *reconcileInvoices) Handle(ctx context.Context) {
	log.Logger.Info("starting invoice reconciliation job")

	count, err := h.svc.Invoice.Reconcile(ctx)
	if err != nil {
		log.Logger.Error("invoice reconciliation failed", zap.Error(err))
		return
	}

	log.Logger.Info("invoice reconciliation completed", zap.Int64("reconciled", count))
}
