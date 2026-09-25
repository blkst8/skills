// Package http contains the Echo HTTP server of the application.
//
// Import it as httpserver because the package name collides with the
// standard library net/http package.
package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/blkst8/client-service/internal/app"
	"github.com/blkst8/client-service/internal/config"
	"github.com/blkst8/client-service/internal/http/handlers"
	"github.com/blkst8/client-service/internal/http/middlewares"
	"github.com/blkst8/client-service/internal/log"
)

// Server wraps the Echo engine and exposes Serve and Shutdown.
type Server struct {
	echo *echo.Echo
}

// NewServer builds the Echo engine, registers middlewares and routes, and
// returns the server. All handlers are created once from the given service
// bundle.
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
	{
		api.POST("/clients", h.CreateClient)
		api.GET("/clients/:id", h.GetClient)
	}

	return &Server{echo: e}
}

// Serve starts the HTTP server on a goroutine using the configured timeouts.
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
		if err := s.echo.StartServer(srv); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatal("failed to start server", zap.Error(err))
		}
	}()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
