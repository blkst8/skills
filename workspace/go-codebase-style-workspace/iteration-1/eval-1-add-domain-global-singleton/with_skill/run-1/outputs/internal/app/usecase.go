package app

import (
	"yourproject/internal/usecase"
)

// Usecase holds all usecase instances. Usecases carry the big business logic;
// they are the only layer HTTP handlers and worker/task handlers talk to.
type Usecase struct {
	Client usecase.ClientUsecase
	// Add other usecases here
}

// WithUsecase constructs all usecases into the global singleton A.
// Must be called after WithRepository.
func WithUsecase() {
	A.Usecase = &Usecase{
		Client: usecase.NewClientUsecase(A.Repository.Client),
	}
}
