// Package http sets up the HTTP server, its middlewares and routes.
//
// Import this package with the alias httpserver to avoid clashing with the
// standard library net/http package.
package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/app"
	"github.com/blkst8/invoice-service/internal/config"
	"github.com/blkst8/invoice-service/internal/http/handlers"
	"github.com/blkst8/invoice-service/internal/http/middlewares"
	"github.com/blkst8/invoice-service/internal/log"
)

// Server wraps the Echo instance used to serve the API.
type Server struct {
	echo *echo.Echo
}

// NewServer builds the Echo instance, registers the middlewares and wires all
// routes to the handlers built from the given service bundle.
func NewServer(svc *app.Service) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middlewares.ZapLogger(log.Logger, "/healthz", "/metrics"))
	e.Use(middleware.CORS())

	h := handlers.New(svc)

	e.GET("/healthz", handlers.Healthz)
	e.GET("/metrics", handlers.Metrics)

	api := e.Group("/api/v1")
	if secret := config.C.HTTPServer.JWTSecret; secret != "" {
		api.Use(middlewares.JWTAuthentication(secret))
	}
	{
		api.POST("/invoices", h.CreateInvoice)
		api.GET("/invoices", h.ListInvoices)
		api.GET("/invoices/:id", h.GetInvoice)
	}

	return &Server{echo: e}
}

// Serve starts the HTTP server in a background goroutine.
func (s *Server) Serve() {
	cfg := config.C.HTTPServer

	srv := &http.Server{
		Addr:              cfg.Listen,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		if err := s.echo.StartServer(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("failed to start server", zap.Error(err))
		}
	}()
}

// Shutdown gracefully drains in-flight requests and stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
