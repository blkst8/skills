// Package httpserver sets up the Echo HTTP server of the application.
package httpserver

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"yourproject/internal/config"
	"yourproject/internal/http/handlers"
	"yourproject/internal/http/middlewares"
	"yourproject/internal/log"
)

// Server wraps the Echo instance of the application.
type Server struct {
	echo *echo.Echo
}

// NewServer creates the Echo server and wires all routes. In the global
// singleton pattern handlers are plain functions that reach the services
// through app.A.
func NewServer() *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middlewares.ZapLogger(log.Logger, "/healthz", "/metrics"))
	e.Use(middleware.CORS())

	e.GET("/healthz", handlers.Healthz)
	e.GET("/metrics", handlers.Metrics)

	api := e.Group("/api/v1")
	api.Use(middlewares.JWTAuthentication())
	{
		api.POST("/clients", handlers.CreateClient)
		api.GET("/clients/:id", handlers.GetClient)
		api.PUT("/clients/:id", handlers.UpdateClient)
		api.DELETE("/clients/:id", handlers.DeleteClient)
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
