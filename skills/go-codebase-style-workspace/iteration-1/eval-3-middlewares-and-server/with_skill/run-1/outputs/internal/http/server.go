package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/config"
	"yourproject/internal/http/handlers"
	"yourproject/internal/http/middlewares"
	"yourproject/internal/log"
)

// Server wraps the Echo instance and its lifecycle.
type Server struct {
	echo *echo.Echo
}

// NewServer builds the Echo server: global middleware first (recover,
// request ID, zap request logging, CORS), then the public routes
// (/healthz, /metrics), then the JWT-protected /api/v1 group.
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
	api.Use(middlewares.JWTAuthentication(config.C.JWT.Secret))
	{
		api.POST("/clients", h.CreateClient)
		api.GET("/clients/:id", h.GetClient)
		api.PUT("/clients/:id", h.UpdateClient)
		api.DELETE("/clients/:id", h.DeleteClient)
	}

	return &Server{echo: e}
}

// Serve starts the HTTP server on a background goroutine using the timeouts
// from config.
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

// Shutdown gracefully drains in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
