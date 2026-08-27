// Package http provides the Echo HTTP server of the application.
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

// Server wraps the Echo instance.
type Server struct {
	echo *echo.Echo
}

// NewServer builds the Echo server and wires all routes from the usecase bundle.
func NewServer(uc *app.Usecase) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middlewares.ZapLogger(log.Logger, "/healthz"))
	e.Use(middleware.CORS())

	h := handlers.New(uc)

	e.GET("/healthz", handlers.Healthz)

	api := e.Group("/api/v1")
	{
		api.POST("/clients", h.CreateClient)
		api.GET("/clients/:id", h.GetClient)
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
		if err := s.echo.StartServer(srv); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatal("failed to start server", zap.Error(err))
		}
	}()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
