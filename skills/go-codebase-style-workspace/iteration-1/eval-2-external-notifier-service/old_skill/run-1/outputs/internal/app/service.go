package app

import (
	"github.com/blkst8/client-service/internal/config"
	"github.com/blkst8/client-service/internal/services"
	"github.com/blkst8/client-service/internal/services/notifier"
)

// Service holds all service instances and shared infrastructure clients
// (notifier, Redis, tracing, etc.) needed across the application.
type Service struct {
	Client   services.ClientService
	Notifier notifier.Notifier
	// Add other services and shared infrastructure clients here
}

// WithServices constructs the notifier client and all services, and returns
// the bundle. The notifier is a shared infrastructure client: it is kept on
// the Service bundle and injected into the services that use it.
func WithServices(repo *Repository) *Service {
	tgNotifier := notifier.New(config.C.Telegram)

	return &Service{
		Notifier: tgNotifier,
		Client:   services.NewClientService(repo.Client, tgNotifier),
	}
}
