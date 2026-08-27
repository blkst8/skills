package app

import (
	"yourproject/internal/usecase"
)

// Service holds all service instances and shared infrastructure clients
// needed across the application.
type Service struct {
	Client usecase.Client
	// Add other services and infra clients (Redis, tracing, ...) here
}

// WithService constructs all services and stores them on the application
// singleton.
func WithService() {
	A.Service = &Service{
		Client: usecase.NewClientUsecase(A.Repository.Client),
	}
}
