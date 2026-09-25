// Package app is a global application object.
//
// It sets up the application and provides access to its dependencies through
// the package-level singleton A. Handlers, workers and commands reach the
// database, repositories and services via app.A.
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
)

// application is the main application struct that holds all the dependencies.
type application struct {
	Database   *sqlx.DB
	Repository *Repository
	Service    *Service

	Ctx    context.Context
	cancel context.CancelFunc
}

// A is the singleton instance of application.
var A *application

func init() {
	A = &application{}
}

// WithGracefulShutdown installs a signal-aware context on the application.
func WithGracefulShutdown() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	A.Ctx = ctx
	A.cancel = cancel
}

// Wait blocks until the application context is cancelled (SIGINT/SIGTERM)
// and releases the signal handler.
func Wait() {
	<-A.Ctx.Done()
	A.cancel()
}
