// Package services contains the business logic layer.
package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/blkst8/invoice-service/internal/models"
	"github.com/blkst8/invoice-service/internal/repository"
)

// ErrInvalidInvoice is returned when an invoice request fails validation.
var ErrInvalidInvoice = errors.New("invalid invoice request")

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

// CreateInvoiceRequest carries the fields needed to create an invoice.
type CreateInvoiceRequest struct {
	CustomerID  int64
	AmountCents int64
	Currency    string
	DueAt       time.Time
}

// InvoiceService is the business logic interface for invoices.
type InvoiceService interface {
	Create(ctx context.Context, req CreateInvoiceRequest) (*models.Invoice, error)
	Get(ctx context.Context, id int64) (*models.Invoice, error)
	List(ctx context.Context, limit, offset int) ([]models.Invoice, error)
	Reconcile(ctx context.Context) (int64, error)
}

type invoiceService struct {
	// db is the shared database handle, available for transactional methods.
	db   *sqlx.DB
	repo repository.Invoice
}

// NewInvoiceService returns the invoice business logic backed by the given
// database handle and repository.
func NewInvoiceService(db *sqlx.DB, repo repository.Invoice) InvoiceService {
	return &invoiceService{db: db, repo: repo}
}

func (s *invoiceService) Create(ctx context.Context, req CreateInvoiceRequest) (*models.Invoice, error) {
	if req.CustomerID <= 0 || req.AmountCents <= 0 || req.DueAt.IsZero() {
		return nil, ErrInvalidInvoice
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	now := time.Now().UTC()
	inv := models.Invoice{
		Number:      generateInvoiceNumber(now),
		CustomerID:  req.CustomerID,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Status:      models.InvoiceStatusPending,
		DueAt:       req.DueAt,
	}

	id, err := s.repo.Create(ctx, inv)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}
	inv.ID = id

	return &inv, nil
}

func (s *invoiceService) Get(ctx context.Context, id int64) (*models.Invoice, error) {
	return s.repo.Get(ctx, id)
}

func (s *invoiceService) List(ctx context.Context, limit, offset int) ([]models.Invoice, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

// Reconcile marks every pending invoice whose due date has passed as overdue
// and returns how many invoices were reconciled.
func (s *invoiceService) Reconcile(ctx context.Context) (int64, error) {
	count, err := s.repo.MarkOverdue(ctx, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to reconcile invoices: %w", err)
	}

	return count, nil
}

// generateInvoiceNumber builds a human-readable invoice number for the given
// timestamp, suffixed with a random component to reduce collisions.
func generateInvoiceNumber(now time.Time) string {
	return fmt.Sprintf("INV-%s-%04d", now.Format("20060102150405"), rand.IntN(10000))
}
