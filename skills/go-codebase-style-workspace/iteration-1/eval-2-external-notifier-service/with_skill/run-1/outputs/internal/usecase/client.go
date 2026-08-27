// Package usecase contains the big business logic and flows. Handlers and job
// handlers call usecases; usecases combine repositories (data access) with
// service clients (external dependencies) and decide what is fatal.
package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/log"
	"github.com/blkst8/client-service/internal/models"
	"github.com/blkst8/client-service/internal/repository"
	"github.com/blkst8/client-service/internal/service/notifier"
)

// CreateClientRequest is the input of the client creation flow.
type CreateClientRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	TelegramID string `json:"telegram_id"`
}

// ClientUsecase holds the client business flows.
type ClientUsecase interface {
	Create(ctx context.Context, req CreateClientRequest) (*models.Client, error)
	Get(ctx context.Context, id uint32) (*models.Client, error)
}

// clientUsecase combines the client repository with external service clients.
type clientUsecase struct {
	repo     repository.Client
	notifier notifier.Notifier
}

// NewClientUsecase returns a ClientUsecase wired to the given dependencies.
func NewClientUsecase(repo repository.Client, notifier notifier.Notifier) ClientUsecase {
	return &clientUsecase{
		repo:     repo,
		notifier: notifier,
	}
}

// Create persists a new client and sends a Telegram welcome message.
//
// The welcome notification is a best-effort side effect: when the notifier
// fails (e.g. Telegram is down) the error is logged and the flow continues,
// so client creation still succeeds. Only persistence failures are fatal.
func (u *clientUsecase) Create(ctx context.Context, req CreateClientRequest) (*models.Client, error) {
	client := models.Client{Name: req.Name, Email: req.Email, TelegramID: req.TelegramID}

	id, err := u.repo.Create(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	client.ID = id

	if req.TelegramID != "" {
		msg := fmt.Sprintf("Welcome, %s! Your account is ready.", req.Name)
		if err := u.notifier.Send(ctx, req.TelegramID, msg); err != nil {
			log.Logger.Error("failed to send welcome notification",
				zap.Error(err),
				zap.Uint32("client_id", client.ID),
				zap.String("telegram_id", req.TelegramID),
			)
			// Non-fatal: continue the flow.
		}
	}

	return &client, nil
}

// Get returns the client with the given id.
func (u *clientUsecase) Get(ctx context.Context, id uint32) (*models.Client, error) {
	client, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return client, nil
}
