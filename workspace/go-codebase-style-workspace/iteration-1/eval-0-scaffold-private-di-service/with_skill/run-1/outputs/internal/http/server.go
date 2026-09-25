// Package http sets up the Echo HTTP server: middleware, routes and
// lifecycle (Serve/Shutdown). Package name is http; importers alias it as
// httpserver.
package http

import (
	"context"
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

// Server wraps the Echo engine.
type Server struct {
	echo *echo.Echo
}

// NewServer builds the Echo engine, installs middleware and wires all routes
// from a single handlers bundle constructed over the usecase bundle.
func NewServer(uc *app.Usecase) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middlewares.ZapLogger(log.Logger, "/healthz", "/metrics"))
	e.Use(middleware.CORS())

	h := handlers.New(uc)

	e.GET("/healthz", handlers.Healthz)
	e.GET("/metrics", handlers.Metrics)

	api := e.Group("/api/v1")
	api.Use(middlewares.JWTAuthentication())
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

	log.Logger.Info("http server listening", zap.String("listen", cfg.Listen))

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

// Shutdown gracefully drains in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
