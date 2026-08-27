// Package usecase contains the big business logic: multi-step flows and
// cross-domain orchestration. HTTP handlers and worker/task handlers call
// usecases only — usecases combine repositories (data access) with service
// clients (external dependencies). One file per domain.
package usecase

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/log"
	"github.com/blkst8/invoice-service/internal/metrics"
	"github.com/blkst8/invoice-service/internal/models"
	"github.com/blkst8/invoice-service/internal/repository"
	"github.com/blkst8/invoice-service/internal/service/notifier"
)

// Invoice status values used by the reconcile flow.
const (
	StatusPending = "pending"
	StatusOverdue = "overdue"
)

// Request/response types owned by the usecase
type CreateInvoiceRequest struct {
	CustomerID uint32    `json:"customer_id"`
	Number     string    `json:"number"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	DueDate    time.Time `json:"due_date"`
}

type ReconcileResult struct {
	Checked  int  `json:"checked"`
	Overdue  int  `json:"overdue"`
	Notified bool `json:"notified"`
}

// Interface definition
type InvoiceUsecase interface {
	Create(ctx context.Context, req CreateInvoiceRequest) (*models.Invoice, error)
	Get(ctx context.Context, id uint32) (*models.Invoice, error)
	List(ctx context.Context) ([]models.Invoice, error)
	Reconcile(ctx context.Context) (*ReconcileResult, error)
}

// Implementation struct (private)
type invoiceUsecase struct {
	repo     repository.Invoice
	notifier notifier.Notifier
}

// Constructor
func NewInvoiceUsecase(repo repository.Invoice, n notifier.Notifier) InvoiceUsecase {
	return &invoiceUsecase{
		repo:     repo,
		notifier: n,
	}
}

// Create persists a new invoice via the repository and then fires a
// "created" notification. The notification is a non-fatal side effect:
// failures are logged and the flow continues.
func (u *invoiceUsecase) Create(ctx context.Context, req CreateInvoiceRequest) (*models.Invoice, error) {
	if req.Currency == "" {
		req.Currency = "USD"
	}

	invoice := models.Invoice{
		CustomerID: req.CustomerID,
		Number:     req.Number,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     StatusPending,
		DueDate:    req.DueDate,
	}

	id, err := u.repo.Create(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}
	invoice.ID = id

	if err := u.notifier.Send(ctx, "invoice-created", fmt.Sprintf("invoice %s created for customer %d", invoice.Number, invoice.CustomerID)); err != nil {
		log.Logger.Error("failed to send invoice created notification", zap.Error(err))
		// Non-fatal: continue the flow
	}

	return &invoice, nil
}

func (u *invoiceUsecase) Get(ctx context.Context, id uint32) (*models.Invoice, error) {
	invoice, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return invoice, nil
}

func (u *invoiceUsecase) List(ctx context.Context) ([]models.Invoice, error) {
	invoices, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

// Reconcile marks pending invoices whose due date has passed as overdue and
// notifies ops via the webhook notifier when any were found. Individual
// update failures are logged and retried on the next run; only a failure to
// list invoices is fatal.
func (u *invoiceUsecase) Reconcile(ctx context.Context) (*ReconcileResult, error) {
	pending, err := u.repo.ListByStatus(ctx, StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending invoices: %w", err)
	}

	result := &ReconcileResult{Checked: len(pending)}
	now := time.Now()

	for _, invoice := range pending {
		if invoice.DueDate.After(now) {
			continue
		}

		if err := u.repo.UpdateStatus(ctx, invoice.ID, StatusOverdue); err != nil {
			log.Logger.Error("failed to mark invoice overdue",
				zap.Uint32("invoice_id", invoice.ID),
				zap.Error(err),
			)
			// Log-and-continue: the next reconcile run will retry
			continue
		}

		result.Overdue++
		metrics.ReconcileOverdue.Inc()
	}

	metrics.ReconcileRuns.Inc()
	metrics.ReconcileChecked.Add(float64(result.Checked))

	if result.Overdue > 0 {
		message := fmt.Sprintf("invoice reconciliation: %d of %d pending invoices marked overdue", result.Overdue, result.Checked)
		if err := u.notifier.Send(ctx, "invoice-reconcile", message); err != nil {
			log.Logger.Error("failed to send reconciliation notification", zap.Error(err))
			// Non-fatal: reconciliation itself succeeded
		} else {
			result.Notified = true
		}
	}

	return result, nil
}
