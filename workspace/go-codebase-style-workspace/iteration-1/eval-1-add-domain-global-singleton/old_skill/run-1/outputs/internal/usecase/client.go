// Package usecase contains the application (business logic) layer.
//
// Usecases orchestrate repositories, own input validation and business
// rules, and are the only layer the HTTP handlers talk to.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"go.uber.org/zap"

	"yourproject/internal/log"
	"yourproject/internal/models"
	"yourproject/internal/repository"
)

// ErrInvalidClientInput is returned when a usecase input fails validation.
// HTTP handlers map it to a 400 response.
var ErrInvalidClientInput = errors.New("invalid client input")

// CreateClientInput carries the payload for the client create flow.
type CreateClientInput struct {
	Name  string
	Email string
}

// Client is the business logic interface for client management.
type Client interface {
	Create(ctx context.Context, input CreateClientInput) (*models.Client, error)
	Get(ctx context.Context, id uint32) (*models.Client, error)
	Update(ctx context.Context, client models.Client) error
	Delete(ctx context.Context, id uint32) error
}

// clientUsecase is the private implementation of the Client usecase.
type clientUsecase struct {
	repo repository.Client
}

// NewClientUsecase builds a new client usecase on top of the given repository.
func NewClientUsecase(repo repository.Client) Client {
	return &clientUsecase{
		repo: repo,
	}
}

// Create validates the input, stores the new client and returns the stored
// record with its generated ID and timestamps.
func (u *clientUsecase) Create(ctx context.Context, input CreateClientInput) (*models.Client, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidClientInput)
	}
	if input.Email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidClientInput)
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return nil, fmt.Errorf("%w: email is not a valid address", ErrInvalidClientInput)
	}

	client := models.Client{
		Name:  input.Name,
		Email: input.Email,
	}

	id, err := u.repo.Create(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	log.Logger.Info("client created", zap.Uint32("id", id))

	created, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get created client: %w", err)
	}

	return created, nil
}

// Get returns a single client by ID.
func (u *clientUsecase) Get(ctx context.Context, id uint32) (*models.Client, error) {
	return u.repo.Get(ctx, id)
}

// Update validates the payload and modifies an existing client.
func (u *clientUsecase) Update(ctx context.Context, client models.Client) error {
	if client.ID == 0 {
		return fmt.Errorf("%w: id is required", ErrInvalidClientInput)
	}
	if client.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidClientInput)
	}
	if client.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidClientInput)
	}
	if _, err := mail.ParseAddress(client.Email); err != nil {
		return fmt.Errorf("%w: email is not a valid address", ErrInvalidClientInput)
	}

	return u.repo.Update(ctx, client)
}

// Delete removes a client by ID.
func (u *clientUsecase) Delete(ctx context.Context, id uint32) error {
	return u.repo.Delete(ctx, id)
}
