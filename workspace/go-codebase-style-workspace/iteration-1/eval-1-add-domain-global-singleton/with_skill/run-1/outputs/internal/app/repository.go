package app

import (
	"yourproject/internal/repository"
)

// Repository holds all repository instances for the application.
type Repository struct {
	Client repository.Client
	// Add other repositories here
}

// WithRepository constructs all repositories into the global singleton A.
// Must be called after WithDatabase.
func WithRepository() {
	A.Repository = &Repository{
		Client: repository.NewClientRepository(A.Database),
	}
}
