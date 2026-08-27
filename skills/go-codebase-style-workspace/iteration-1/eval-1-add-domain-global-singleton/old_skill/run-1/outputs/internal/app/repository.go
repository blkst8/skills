package app

import (
	"yourproject/internal/repository"
)

// Repository holds all repository instances for the application.
type Repository struct {
	Client repository.Client
	// Add other repositories here
}

// WithRepository constructs all repositories and stores them on the
// application singleton.
func WithRepository() {
	A.Repository = &Repository{
		Client: repository.NewClientRepository(A.Database),
	}
}
