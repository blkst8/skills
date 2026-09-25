// Package handlers contains the interval-based job handlers. Like HTTP
// handlers, job handlers call usecases only.
package handlers

import (
	"context"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/app"
	"github.com/blkst8/invoice-service/internal/log"
)

type reconcileInvoices struct {
	uc *app.Usecase
}

func NewReconcileInvoices(uc *app.Usecase) *reconcileInvoices {
	return &reconcileInvoices{uc: uc}
}

// Handle runs one invoice reconciliation pass.
func (h *reconcileInvoices) Handle(ctx context.Context) {
	log.Logger.Info("starting invoice reconciliation job")

	result, err := h.uc.Invoice.Reconcile(ctx)
	if err != nil {
		log.Logger.Error("invoice reconciliation failed", zap.Error(err))
		return
	}

	log.Logger.Info("invoice reconciliation completed",
		zap.Int("checked", result.Checked),
		zap.Int("overdue", result.Overdue),
		zap.Bool("notified", result.Notified),
	)
}
