// Package services implements the business logic of the application.
//
// Services depend on repository interfaces and infrastructure clients
// (notifier, ...) that are injected through constructors, never on globals.
package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/log"
	"github.com/blkst8/client-service/internal/models"
	"github.com/blkst8/client-service/internal/repository"
	"github.com/blkst8/client-service/internal/services/notifier"
)

// welcomeMessageFormat is the Telegram welcome message sent to new clients.
const welcomeMessageFormat = "Welcome aboard, %s! Your account has been created successfully."

// ClientService defines the client-related business operations.
type ClientService interface {
	// Create stores a new client and sends a Telegram welcome message.
	Create(ctx context.Context, req CreateClientRequest) (*models.Client, error)
	// Get returns the client with the given ID.
	Get(ctx context.Context, id uint32) (*models.Client, error)
}

// CreateClientRequest holds the input for creating a new client.
type CreateClientRequest struct {
	Name       string
	Email      string
	TelegramID string
}

// clientService is the default ClientService implementation.
type clientService struct {
	clients  repository.Client
	notifier notifier.Notifier
}

// NewClientService returns a ClientService backed by the given repository and
// notifier.
func NewClientService(clients repository.Client, notifier notifier.Notifier) ClientService {
	return &clientService{
		clients:  clients,
		notifier: notifier,
	}
}

// Create stores the client and then sends a Telegram welcome message to the
// client's chat.
//
// The welcome message is best effort: when the notifier fails (for example
// because Telegram is down) the error is logged and client creation still
// succeeds. Creation is committed before the message is sent, so a notifier
// failure can never roll back or fail the sign-up.
func (c *clientService) Create(ctx context.Context, req CreateClientRequest) (*models.Client, error) {
	client := models.Client{
		Name:       req.Name,
		Email:      req.Email,
		TelegramID: req.TelegramID,
	}

	if err := c.clients.Create(ctx, &client); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	c.sendWelcomeMessage(ctx, client)

	return &client, nil
}

// Get returns the client with the given ID.
func (c *clientService) Get(ctx context.Context, id uint32) (*models.Client, error) {
	client, err := c.clients.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return client, nil
}

// sendWelcomeMessage notifies the client on Telegram. Failures are logged and
// never propagated to the caller so they cannot fail client creation.
func (c *clientService) sendWelcomeMessage(ctx context.Context, client models.Client) {
	if client.TelegramID == "" {
		return
	}

	message := fmt.Sprintf(welcomeMessageFormat, client.Name)
	if err := c.notifier.SendMessage(ctx, client.TelegramID, message); err != nil {
		log.Logger.Error("failed to send telegram welcome message",
			zap.Error(err),
			zap.Uint32("client_id", client.ID),
			zap.String("telegram_id", client.TelegramID),
		)
	}
}
