// Package handlers provides HTTP request handlers for the application.
//
// Each handler is responsible for:
//   - Binding and validating request input
//   - Calling the matching usecase
//   - Logging errors
//   - Mapping sentinel errors to HTTP status codes
//
// Handlers never call repositories or services directly. With private
// dependency injection all handlers live as methods on the Handlers struct.
package handlers

import "github.com/blkst8/invoice-service/internal/app"

// Handlers holds the usecase bundle and exposes all handler methods.
type Handlers struct {
	uc *app.Usecase
}

// New returns a Handlers wired to the given usecase bundle.
func New(uc *app.Usecase) *Handlers {
	return &Handlers{uc: uc}
}
