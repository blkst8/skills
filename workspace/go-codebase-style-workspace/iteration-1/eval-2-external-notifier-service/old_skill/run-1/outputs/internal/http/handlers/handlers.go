// Package handlers provides HTTP request handlers for the application.
//
// Each handler is responsible for:
//   - Validating request input
//   - Calling appropriate service methods
//   - Formatting response output
//   - Logging errors
//
// One file per handler action; request and response types are local to their
// file. All handlers are methods on the Handlers struct, which holds the
// service bundle injected in New.
package handlers

import "github.com/blkst8/client-service/internal/app"

// Handlers holds the service bundle and exposes all handler methods.
type Handlers struct {
	svc *app.Service
}

// New returns the handlers wired to the given service bundle.
func New(svc *app.Service) *Handlers {
	return &Handlers{svc: svc}
}
