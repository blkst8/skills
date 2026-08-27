package usecase

import (
	"context"
	"fmt"

	"yourproject/internal/models"
	"yourproject/internal/repository"
)

// Request types owned by the usecase
type CreateClientRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateClientRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Interface definition
type ClientUsecase interface {
	Create(ctx context.Context, req CreateClientRequest) (*models.Client, error)
	Get(ctx context.Context, id uint32) (*models.Client, error)
	Update(ctx context.Context, id uint32, req UpdateClientRequest) (*models.Client, error)
	Delete(ctx context.Context, id uint32) error
}

// Implementation struct (private)
type clientUsecase struct {
	repo repository.Client
}

// Constructor
func NewClientUsecase(repo repository.Client) ClientUsecase {
	return &clientUsecase{
		repo: repo,
	}
}

// Create persists a new client and returns it with the generated ID.
func (u *clientUsecase) Create(ctx context.Context, req CreateClientRequest) (*models.Client, error) {
	client := &models.Client{Name: req.Name, Email: req.Email}
	if err := u.repo.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

func (u *clientUsecase) Get(ctx context.Context, id uint32) (*models.Client, error) {
	client, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return client, nil
}

// Update writes the new values and reads the row back so the response
// carries the stored timestamps.
func (u *clientUsecase) Update(ctx context.Context, id uint32, req UpdateClientRequest) (*models.Client, error) {
	client := models.Client{ID: id, Name: req.Name, Email: req.Email}
	if err := u.repo.Update(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	updated, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return updated, nil
}

func (u *clientUsecase) Delete(ctx context.Context, id uint32) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}
