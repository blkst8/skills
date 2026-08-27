// Package notifier provides clients for delivering messages to users through
// external systems such as the Telegram Bot API.
//
// The Notifier interface decouples business logic from the delivery
// mechanism, which keeps the client-creation flow resilient: when the
// external system is down the caller only has to log the error.
package notifier

import (
	"context"

	"github.com/blkst8/client-service/internal/config"
)

// Notifier sends a text message to a chat identified by chatID.
//
// Implementations must be safe for concurrent use and must return an error
// instead of panicking when the external system is unreachable.
type Notifier interface {
	SendMessage(ctx context.Context, chatID string, text string) error
}

// New returns the Notifier selected by the configuration.
//
// When Telegram notifications are disabled it returns a no-op notifier, so
// neither the service layer nor the wiring code needs to check whether
// notifications are enabled.
func New(cfg config.Telegram) Notifier {
	if !cfg.Enabled {
		return &noop{}
	}

	return newTelegram(cfg)
}

// noop is a Notifier that discards every message. It is used when Telegram
// notifications are disabled in the configuration.
type noop struct{}

// SendMessage discards the message and always succeeds.
func (n *noop) SendMessage(_ context.Context, _ string, _ string) error {
	return nil
}
