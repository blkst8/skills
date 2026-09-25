package workers

import (
	"context"

	"github.com/blkst8/invoice-service/internal/app"
	"github.com/blkst8/invoice-service/internal/log"
)

type reconcileInvoicesTask struct {
	uc *app.Usecase
}

func NewReconcileInvoicesTask(uc *app.Usecase) *reconcileInvoicesTask {
	return &reconcileInvoicesTask{uc: uc}
}

func (t *reconcileInvoicesTask) Name() string { return "reconcile_invoices" }

func (t *reconcileInvoicesTask) Execute(ctx context.Context) error {
	log.Logger.Info("running invoice reconciliation task")
	_, err := t.uc.Invoice.Reconcile(ctx)
	return err
}
