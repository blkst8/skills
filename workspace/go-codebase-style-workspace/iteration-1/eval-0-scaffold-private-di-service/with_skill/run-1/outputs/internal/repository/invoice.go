// Package repository provides sqlx-based data access. Repositories own ALL
// SQL; they translate driver errors (sql.ErrNoRows, MySQL duplicate entry)
// into typed sentinel errors so upper layers never touch database/sql
// internals.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/blkst8/invoice-service/internal/models"
)

var (
	ErrInvoiceNotFound    = errors.New("invoice not found")
	ErrInvoiceNumberTaken = errors.New("invoice number already taken")
)

// Interface definition
type Invoice interface {
	Create(ctx context.Context, invoice models.Invoice) (uint32, error)
	Get(ctx context.Context, id uint32) (*models.Invoice, error)
	List(ctx context.Context) ([]models.Invoice, error)
	ListByStatus(ctx context.Context, status string) ([]models.Invoice, error)
	UpdateStatus(ctx context.Context, id uint32, status string) error
}

// Implementation struct (private)
type invoice struct {
	db *sqlx.DB
}

// Constructor returns the interface
func NewInvoiceRepository(db *sqlx.DB) Invoice {
	return &invoice{
		db: db,
	}
}

func (r *invoice) Create(ctx context.Context, invoice models.Invoice) (uint32, error) {
	query := `INSERT INTO invoices (customer_id, number, amount, currency, status, due_date)
              VALUES (:customer_id, :number, :amount, :currency, :status, :due_date)`
	result, err := r.db.NamedExecContext(ctx, query, &invoice)
	if err != nil {
		if isDuplicateEntry(err) {
			return 0, ErrInvoiceNumberTaken
		}
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint32(id), nil
}

func (r *invoice) Get(ctx context.Context, id uint32) (*models.Invoice, error) {
	query := `SELECT * FROM invoices WHERE id = ?`
	var invoice models.Invoice
	err := r.db.GetContext(ctx, &invoice, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *invoice) List(ctx context.Context) ([]models.Invoice, error) {
	query := `SELECT * FROM invoices ORDER BY id`
	var invoices []models.Invoice
	err := r.db.SelectContext(ctx, &invoices, query)
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *invoice) ListByStatus(ctx context.Context, status string) ([]models.Invoice, error) {
	query := `SELECT * FROM invoices WHERE status = ? ORDER BY due_date`
	var invoices []models.Invoice
	err := r.db.SelectContext(ctx, &invoices, query, status)
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *invoice) UpdateStatus(ctx context.Context, id uint32, status string) error {
	query := `UPDATE invoices SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// isDuplicateEntry reports whether err is a MySQL duplicate-key error.
func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
