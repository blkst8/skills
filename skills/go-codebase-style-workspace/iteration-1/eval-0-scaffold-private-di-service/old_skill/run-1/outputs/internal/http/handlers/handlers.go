// Package handlers provides HTTP request handlers for the application.
//
// With the private dependency injection pattern every handler is a method on
// the single Handlers struct; each action lives in its own file.
package handlers

import "github.com/blkst8/invoice-service/internal/app"

// Handlers holds the service bundle and exposes all handler methods.
type Handlers struct {
	svc *app.Service
}

// New returns a Handlers wired to the given service bundle.
func New(svc *app.Service) *Handlers {
	return &Handlers{svc: svc}
}
