package app

import (
	"github.com/jmoiron/sqlx"

	"github.com/blkst8/client-service/internal/repository"
)

// Repository holds all repository instances for the application.
type Repository struct {
	Client repository.Client
	// Add other repositories here
}

// WithRepository constructs all repositories and returns the bundle.
func WithRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Client: repository.NewClientRepository(db),
	}
}
