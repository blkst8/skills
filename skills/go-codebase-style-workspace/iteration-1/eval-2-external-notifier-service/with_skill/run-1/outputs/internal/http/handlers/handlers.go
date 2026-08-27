// Package handlers provides HTTP request handlers for the application.
//
// Each handler is responsible for:
//   - Validating request input
//   - Calling the appropriate usecase
//   - Mapping sentinel errors to HTTP status codes
//   - Formatting response output and logging errors
package handlers

import "github.com/blkst8/client-service/internal/app"

// Handlers holds the usecase bundle and exposes all handler methods.
type Handlers struct {
	uc *app.Usecase
}

// New returns a Handlers bound to the given usecase bundle.
func New(uc *app.Usecase) *Handlers {
	return &Handlers{uc: uc}
}
