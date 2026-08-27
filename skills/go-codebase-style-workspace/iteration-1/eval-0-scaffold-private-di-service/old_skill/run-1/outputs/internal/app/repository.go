package app

import (
	"github.com/jmoiron/sqlx"

	"github.com/blkst8/invoice-service/internal/repository"
)

// Repository holds all repository instances for the application.
type Repository struct {
	Invoice repository.Invoice
}

// WithRepository constructs all repositories and returns the bundle.
func WithRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Invoice: repository.NewInvoiceRepository(db),
	}
}
