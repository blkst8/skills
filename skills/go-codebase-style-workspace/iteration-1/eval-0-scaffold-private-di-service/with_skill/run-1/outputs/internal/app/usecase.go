package app

import (
	"github.com/blkst8/invoice-service/internal/usecase"
)

// Usecase holds all usecase instances. Usecases carry the big business logic;
// they are the only layer HTTP handlers and worker/task handlers talk to.
type Usecase struct {
	Invoice usecase.InvoiceUsecase
	// Add other usecases here
}

// WithUsecases constructs all usecases and returns the bundle.
func WithUsecases(repo *Repository, svc *Service) *Usecase {
	return &Usecase{
		Invoice: usecase.NewInvoiceUsecase(repo.Invoice, svc.Notifier),
	}
}
