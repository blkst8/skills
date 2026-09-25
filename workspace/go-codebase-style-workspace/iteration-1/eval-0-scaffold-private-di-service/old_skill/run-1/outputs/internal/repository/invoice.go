// Package repository contains the data access layer.
//
// Every repository is declared as an interface with a private sqlx-backed
// implementation, so services can depend on the interface in tests.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/blkst8/invoice-service/internal/models"
)

// ErrInvoiceNotFound is returned when no invoice matches the given identifier.
var ErrInvoiceNotFound = errors.New("invoice not found")

// Invoice is the data access interface for invoices.
type Invoice interface {
	Create(ctx context.Context, invoice models.Invoice) (int64, error)
	Get(ctx context.Context, id int64) (*models.Invoice, error)
	List(ctx context.Context, limit, offset int) ([]models.Invoice, error)
	MarkOverdue(ctx context.Context, now time.Time) (int64, error)
}

type invoice struct {
	db *sqlx.DB
}

// NewInvoiceRepository returns the sqlx-backed invoice repository.
func NewInvoiceRepository(db *sqlx.DB) Invoice {
	return &invoice{db: db}
}

const invoiceColumns = "id, number, customer_id, amount_cents, currency, status, due_at, created_at, updated_at"

func (i *invoice) Create(ctx context.Context, inv models.Invoice) (int64, error) {
	query := `
		INSERT INTO invoices (number, customer_id, amount_cents, currency, status, due_at)
		VALUES (:number, :customer_id, :amount_cents, :currency, :status, :due_at)`

	result, err := i.db.NamedExecContext(ctx, query, &inv)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (i *invoice) Get(ctx context.Context, id int64) (*models.Invoice, error) {
	query := `SELECT ` + invoiceColumns + ` FROM invoices WHERE id = ?`

	var inv models.Invoice
	if err := i.db.GetContext(ctx, &inv, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}

	return &inv, nil
}

func (i *invoice) List(ctx context.Context, limit, offset int) ([]models.Invoice, error) {
	query := `SELECT ` + invoiceColumns + ` FROM invoices ORDER BY id DESC LIMIT ? OFFSET ?`

	invoices := make([]models.Invoice, 0)
	if err := i.db.SelectContext(ctx, &invoices, query, limit, offset); err != nil {
		return nil, err
	}

	return invoices, nil
}

// MarkOverdue flags every pending invoice whose due date has passed and
// returns how many rows were updated.
func (i *invoice) MarkOverdue(ctx context.Context, now time.Time) (int64, error) {
	query := `
		UPDATE invoices
		SET status = ?, updated_at = ?
		WHERE status = ? AND due_at < ?`

	result, err := i.db.ExecContext(ctx, query,
		models.InvoiceStatusOverdue, now,
		models.InvoiceStatusPending, now,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
