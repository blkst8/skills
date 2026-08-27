package app

import (
	"github.com/blkst8/client-service/internal/config"
	"github.com/blkst8/client-service/internal/service/notifier"
)

// Service holds all clients for external dependencies (third-party APIs,
// mail/SMS providers, message brokers, storage, ...).
type Service struct {
	Notifier notifier.Notifier
	// Add other external service clients here (cache, storage, ...)
}

// WithServices constructs all external service clients and returns the bundle.
func WithServices() *Service {
	return &Service{
		Notifier: notifier.New(config.C.Telegram.Token),
	}
}
