package app

import (
	"github.com/blkst8/client-service/internal/usecase"
)

// Usecase holds all usecase instances. Usecases carry the big business logic;
// they are the only layer HTTP handlers and worker/task handlers talk to.
type Usecase struct {
	Client usecase.ClientUsecase
	// Add other usecases here
}

// WithUsecases constructs all usecases and returns the bundle.
func WithUsecases(repo *Repository, svc *Service) *Usecase {
	return &Usecase{
		Client: usecase.NewClientUsecase(repo.Client, svc.Notifier),
	}
}
