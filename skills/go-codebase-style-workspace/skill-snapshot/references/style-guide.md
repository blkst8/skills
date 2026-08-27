# Go Codebase Style Guide

> A comprehensive style guide for building production-ready Go applications based on proven patterns and best practices.

**Version:** 1.0  
**Last Updated:** December 25, 2025

---

## Table of Contents

- [Project Structure](#project-structure)
- [Architecture Patterns](#architecture-patterns)
- [Code Organization](#code-organization)
- [Naming Conventions](#naming-conventions)
- [Error Handling](#error-handling)
- [Logging](#logging)
- [Configuration Management](#configuration-management)
- [Database Layer](#database-layer)
- [HTTP Layer](#http-layer)
- [Worker/Background Jobs](#workerbackground-jobs)
- [Dependency Management](#dependency-management)
- [Build and Deployment](#build-and-deployment)
- [Documentation](#documentation)

---

## Project Structure

### Standard Directory Layout

```
project/
├── cmd/                        # Command-line interface commands
│   ├── root.go                # Root command (cobra)
│   ├── start.go               # Start server command
│   ├── migrate.go             # Database migration command
│   └── jwt.go                 # Utility commands
├── internal/                   # Private application code
│   ├── app/                   # Global application setup
│   │   ├── app.go             # Application singleton
│   │   ├── db.go              # Database initialization
│   │   ├── repository.go      # Repository initialization
│   │   └── service.go         # Service initialization
│   ├── config/                # Configuration management
│   │   ├── config.go          # Config structures
│   │   └── builtin.go         # Default configurations
│   ├── http/                  # HTTP server layer
│   │   ├── server.go          # Server setup
│   │   ├── handlers/          # HTTP handlers
│   │   └── middlewares/       # HTTP middlewares
│   ├── log/                   # Logging setup
│   │   └── log.go
│   ├── metrics/               # Metrics and monitoring
│   │   ├── business.go        # Business metrics
│   │   └── monitoring.go      # System metrics
│   ├── models/                # Data models
│   ├── repository/            # Data access layer
│   └── worker/                # Background workers
│       ├── worker.go          # Worker framework
│       └── handlers/          # Job handlers
├── pkg/                       # Public library code
│   └── client/                # Reusable client packages
├── migrations/                # Database migrations
│   ├── YYYYMMDDHHMMSS_name.up.sql
│   └── YYYYMMDDHHMMSS_name.down.sql
├── doc/                       # Project documentation
│   ├── features/             # Feature specifications
│   └── architecture/         # Architecture decision records
├── deployments/               # Deployment configurations
│   └── systemd/              # Systemd service files
├── main.go                    # Application entry point
├── go.mod                     # Go module dependencies
├── go.sum                     # Dependency checksums
├── Makefile                   # Build automation
├── config.example.yaml        # Example configuration
└── README.md                  # Project documentation
```

### Key Principles

1. **`internal/` for private code**: Code that should not be imported by other projects
2. **`pkg/` for public libraries**: Reusable packages that can be imported
3. **`cmd/` for executables**: Each subdirectory is a separate executable
4. **Separation of concerns**: Clear boundaries between layers

---

## Architecture Patterns

### Layered Architecture

```
┌─────────────────────────────────────┐
│         HTTP/CLI Layer              │  (cmd/, internal/http/)
├─────────────────────────────────────┤
│         Service Layer               │  (internal/app/service.go)
├─────────────────────────────────────┤
│         Repository Layer            │  (internal/repository/)
├─────────────────────────────────────┤
│         Models Layer                │  (internal/models/)
├─────────────────────────────────────┤
│         Database Layer              │  (internal/app/db.go)
└─────────────────────────────────────┘
```

### Dependency Wiring Patterns

The **Database**, **Repository**, and **Service** layers support two wiring strategies. **Choose one per project** and apply it consistently.

| | Global Singleton | Private Dependencies |
|---|---|---|
| Access | `app.A.*` global | Injected via constructors |
| Testability | Low | High |
| Best for | Simple scripts / small services | Production services, multi-instance, testable code |

---

### Pattern 1: Global Singleton

**Purpose**: Centralized access to application-wide dependencies via a package-level variable.

**File**: `internal/app/app.go`

```go
// Package app is a global application object.
//
// It sets up the application and provides access to the application.
package app

import (
    "context"
    "github.com/jmoiron/sqlx"
)

// application is the main application struct that holds all the dependencies.
type application struct {
    Database     *sqlx.DB
    Repository   *Repository
    Service      *Service

    Ctx        context.Context
    cancelFunc context.CancelFunc
}

// A is the singleton instance of application.
var A *application

func init() {
    A = &application{}
}

// Builder pattern functions — mutate the global A in place.
func WithGracefulShutdown() { /* ... */ }
func WithDatabase()         { /* ... */ }
func WithRepository()       { /* ... */ }
func WithService()          { /* ... */ }
func Wait()                 { /* ... */ }
```

**Usage in `cmd/start.go`**:

```go
func startFunc(_ *cobra.Command, _ []string) {
    app.WithGracefulShutdown()
    app.WithDatabase()
    app.WithRepository()
    app.WithService()

    srv := httpserver.NewServer()
    srv.Serve()

    app.Wait()
}
```

**Handlers access dependencies via the global**:

```go
func CreateClient(ctx echo.Context) error {
    result, err := app.A.Service.Client.Create(ctx.Request().Context(), req)
    // ...
}
```

---

### Pattern 2: Private Dependencies (Dependency Injection)

**Purpose**: Each layer receives its dependencies explicitly. No global state. Suitable for production services requiring testability and clean boundaries.

The `app` package exposes typed structs (`Repository`, `Service`) and constructor functions (`WithRepository`, `WithServices`) that return concrete instances. There is no `application` singleton struct.

#### `internal/app/db.go`

```go
package app

import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "yourproject/internal/config"
    "yourproject/internal/log"
)

// WithDatabase opens and validates the database connection and returns it.
func WithDatabase() *sqlx.DB {
    cfg := config.C.Database

    db, err := sqlx.Open("mysql", cfg.DSN)
    if err != nil {
        log.Logger.Fatal("failed to connect to database", zap.Error(err))
    }

    db.SetMaxOpenConns(cfg.MaxConn)
    db.SetMaxIdleConns(cfg.IdleConn)
    db.SetConnMaxLifetime(cfg.Timeout)

    if err := db.Ping(); err != nil {
        log.Logger.Fatal("failed to ping database", zap.Error(err))
    }

    return db
}
```

#### `internal/app/repository.go`

```go
package app

import (
    "github.com/jmoiron/sqlx"

    "yourproject/internal/repository"
)

// Repository holds all repository instances for the application.
type Repository struct {
    Client repository.Client
    // Add other repositories here
}

// WithRepository constructs all repositories and returns the bundle.
func WithRepository(db *sqlx.DB) *Repository {
    return &Repository{
        Client: repository.NewClientRepository(db),
    }
}
```

#### `internal/app/service.go`

```go
package app

import (
    "github.com/jmoiron/sqlx"

    "yourproject/internal/services"
)

// Service holds all service instances and shared infrastructure clients
// (db, redis, tracing, etc.) needed across the application.
type Service struct {
    Client services.ClientService
    Auth   services.AuthService
    // Add other internal services, Redis clients, tracing providers, etc.
}

// WithServices constructs all services and returns the bundle.
func WithServices(db *sqlx.DB, repo *Repository) *Service {
    return &Service{
        Client: services.NewClientService(db, repo.Client),
        Auth:   services.NewAuthService(db, repo.Client),
    }
}
```

#### `cmd/start.go` — wiring everything together

```go
func startFunc(_ *cobra.Command, _ []string) {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    db   := app.WithDatabase()
    defer db.Close()

    repo := app.WithRepository(db)
    svc  := app.WithServices(db, repo)

    // Start workers
    if config.C.Worker.Enabled {
        syncJob := worker.NewWorker(config.C.Worker.JobsIntervals.SyncDatabases)
        syncJob.RunAsync(workerhandlers.NewSyncDatabases(svc).Handle)
        defer syncJob.Close()
    }

    // Start HTTP server
    srv := httpserver.NewServer(svc)
    srv.Serve()
    defer srv.Shutdown(context.Background())

    <-ctx.Done()
}
```

---

## Code Organization

### Package Organization

#### 1. Handler Package Pattern

**Location**: `internal/http/handlers/`

**One handler per file** named after the action:
- `create_client.go`
- `update_client.go`
- `delete_client.go`
- `get_client.go`

**Global Singleton pattern** — handler is a plain function accessing `app.A`:

```go
package handlers

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"

    "yourproject/internal/app"
    "yourproject/internal/log"
)

type CreateResourceRequest struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}

type CreateResourceResponse struct {
    ID    int      `json:"id"`
    Links []string `json:"links"`
}

func CreateResource(ctx echo.Context) error {
    var request CreateResourceRequest
    if err := ctx.Bind(&request); err != nil {
        log.Logger.Error("failed to bind request", zap.Error(err))
        return err
    }

    result, err := app.A.Service.Client.Create(ctx.Request().Context(), request)
    if err != nil {
        log.Logger.Error("failed to create resource", zap.Error(err))
        return err
    }

    return ctx.JSON(http.StatusOK, CreateResourceResponse{ID: result.ID})
}
```

**Private Dependencies pattern** — all handlers live as methods on a single `Handlers` struct. One constructor, all routes covered:

**File**: `internal/http/handlers/handlers.go`

```go
package handlers

import "yourproject/internal/app"

// Handlers holds the service bundle and exposes all handler methods.
type Handlers struct {
    svc *app.Service
}

func New(svc *app.Service) *Handlers {
    return &Handlers{svc: svc}
}
```

**File**: `internal/http/handlers/create_resource.go`

```go
package handlers

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"

    "yourproject/internal/log"
)

type CreateResourceRequest struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}

type CreateResourceResponse struct {
    ID    int      `json:"id"`
    Links []string `json:"links"`
}

func (h *Handlers) CreateResource(ctx echo.Context) error {
    var request CreateResourceRequest
    if err := ctx.Bind(&request); err != nil {
        log.Logger.Error("failed to bind request", zap.Error(err))
        return err
    }

    result, err := h.svc.Client.Create(ctx.Request().Context(), request)
    if err != nil {
        log.Logger.Error("failed to create resource", zap.Error(err))
        return err
    }

    return ctx.JSON(http.StatusOK, CreateResourceResponse{ID: result.ID})
}
```

#### 2. Repository Package Pattern

**Location**: `internal/repository/`

**Interface-driven design**:
```go
package repository

import (
    "context"
    "github.com/jmoiron/sqlx"
    "yourproject/internal/models"
)

// Interface definition
type Client interface {
    Create(ctx context.Context, client models.Client) error
    Get(ctx context.Context, id string) (*models.Client, error)
    Update(ctx context.Context, client models.Client) error
    Delete(ctx context.Context, id string) error
}

// Implementation struct (private)
type client struct {
    db *sqlx.DB
}

// Constructor
func NewClientRepository(db *sqlx.DB) Client {
    return &client{
        db: db,
    }
}

// Methods
func (c *client) Create(ctx context.Context, client models.Client) error {
    query := `INSERT INTO clients (...) VALUES (...)`
    _, err := c.db.NamedExecContext(ctx, query, &client)
    return err
}
```

**Key Points**:
- Define interface for testability
- Private implementation struct
- Public constructor returning interface
- Use `context.Context` as first parameter
- Use named queries with `sqlx`

#### 3. Models Package Pattern

**Location**: `internal/models/`

```go
package models

import "time"

type Client struct {
    ID         uint32     `db:"id" json:"id"`
    Name       string     `db:"name" json:"name"`
    Email      string     `db:"email" json:"email"`
    CreatedAt  time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt  *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}
```

**Guidelines**:
- Use struct tags for both `db` and `json`
- Use pointers for nullable fields
- Keep models simple (no business logic)

---

## Naming Conventions

### Files

- **Snake case**: `create_client.go`, `jwt_authentication.go`
- **Descriptive names**: Name files after their primary function
- **Test files**: `*_test.go`

### Packages

- **Lowercase, single word**: `handlers`, `models`, `repository`
- **Plural for collections**: `handlers`, `middlewares`
- **Avoid generic names**: Use `httpserver` instead of `server`

### Variables and Functions

```go
// Variables: camelCase for local, PascalCase for exported
var localVariable string
var ExportedVariable string

// Functions: PascalCase for exported, camelCase for private
func PublicFunction() {}
func privateFunction() {}

// Constants: PascalCase or SCREAMING_SNAKE_CASE
const MaxRetries = 3
const DEFAULT_TIMEOUT = 30

// Interfaces: noun or adjective, not -er suffix preferred
type Client interface {}  // Good
type Reader interface {}  // Acceptable
```

### Receivers

- **Short and consistent**: Use 1-2 letter abbreviations
- **Not `this` or `self`**
```go
type Client struct {}

func (c *client) Get() {}  // Good
func (cl *client) Get() {} // OK if 'c' conflicts
func (this *client) Get() {} // Bad
```

---

## Error Handling

### Standard Pattern

```go
// Check errors immediately
result, err := someFunction()
if err != nil {
    log.Logger.Error("descriptive message", zap.Error(err))
    return err
}

// Use errors.New for constants
var ErrClientNotFound = errors.New("client not found")

// Wrap errors with context (if needed)
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}
```

### Repository Error Handling

```go
import (
    "database/sql"
    "errors"
)

var ErrClientNotFound = errors.New("client not found")

func (c *client) Get(ctx context.Context, id string) (*models.Client, error) {
    var client models.Client
    err := c.db.GetContext(ctx, &client, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrClientNotFound
        }
        return nil, err
    }
    return &client, nil
}
```

### HTTP Error Handling

**Global Singleton pattern**:

```go
func Handler(ctx echo.Context) error {
    result, err := app.A.Service.DoSomething()
    if err != nil {
        log.Logger.Error("operation failed", zap.Error(err))

        if errors.Is(err, repository.ErrNotFound) {
            return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
        }

        return err
    }

    return ctx.JSON(http.StatusOK, result)
}
```

**Private Dependencies pattern** — same logic inside a handler struct method:

```go
func (h *myHandler) Handle(ctx echo.Context) error {
    result, err := h.svc.DoSomething()
    if err != nil {
        log.Logger.Error("operation failed", zap.Error(err))

        if errors.Is(err, repository.ErrNotFound) {
            return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
        }

        return err
    }

    return ctx.JSON(http.StatusOK, result)
}
```

---

## Logging

### Setup with Zap

**File**: `internal/log/log.go`

```go
package log

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
    var err error
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    Logger, err = config.Build()
    if err != nil {
        panic(err)
    }
}
```

### Usage Patterns

```go
import (
    "go.uber.org/zap"
    "yourproject/internal/log"
)

// Info level
log.Logger.Info("server started", zap.String("address", ":8080"))

// Error level with error
log.Logger.Error("failed to connect", zap.Error(err))

// With multiple fields
log.Logger.Info("request processed",
    zap.String("method", "POST"),
    zap.String("path", "/api/clients"),
    zap.Int("status", 200),
    zap.Duration("latency", duration),
)

// Structured logging
log.Logger.Debug("processing item",
    zap.String("id", itemID),
    zap.Any("details", item),
)
```

### HTTP Request Logging

Use middleware for automatic request/response logging:

```go
// internal/http/middlewares/logger.go
func ZapLogger(log *zap.Logger, skipURLs ...string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(ctx echo.Context) error {
            start := time.Now()
            err := next(ctx)
            
            // Log with structured fields
            log.Info("request",
                zap.String("method", ctx.Request().Method),
                zap.String("uri", ctx.Request().RequestURI),
                zap.Int("status", ctx.Response().Status),
                zap.Duration("latency", time.Since(start)),
            )
            
            return err
        }
    }
}
```

---

## Configuration Management

### Using Viper

**File**: `internal/config/config.go`

```go
package config

import (
    "github.com/spf13/viper"
    "time"
)

var C *Config

type Config struct {
    Logger     Logger     `yaml:"logger"`
    HTTPServer HTTPServer `yaml:"http_server"`
    Database   Database   `yaml:"database"`
    Worker     Worker     `yaml:"worker"`
}

type Logger struct {
    Level string `yaml:"level"`
}

type HTTPServer struct {
    Listen            string        `yaml:"listen"`
    ReadTimeout       time.Duration `yaml:"read_timeout"`
    WriteTimeout      time.Duration `yaml:"write_timeout"`
    ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
    IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

type Database struct {
    DSN         string        `yaml:"dsn"`
    MaxConn     int           `yaml:"max_conn"`
    IdleConn    int           `yaml:"idle_conn"`
    Timeout     time.Duration `yaml:"timeout"`
}

type Worker struct {
    Enabled       bool          `yaml:"enabled"`
    JobsIntervals JobsIntervals `yaml:"jobs_intervals"`
}

type JobsIntervals struct {
    SyncDatabases time.Duration `yaml:"sync_databases"`
    CleanupJobs   time.Duration `yaml:"cleanup_jobs"`
}

func Load(configPath string) error {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("yaml")
    
    if err := viper.ReadInConfig(); err != nil {
        return err
    }
    
    if err := viper.Unmarshal(&C); err != nil {
        return err
    }
    
    return nil
}
```

### Configuration File (YAML)

**File**: `config.example.yaml`

```yaml
logger:
  level: info

http_server:
  listen: :8080
  read_timeout: 30s
  write_timeout: 30s
  read_header_timeout: 10s
  idle_timeout: 120s

database:
  dsn: "user:pass@tcp(localhost:3306)/dbname?parseTime=true"
  max_conn: 25
  idle_conn: 5
  timeout: 30s

worker:
  enabled: true
  jobs_intervals:
    sync_databases: 5m
    cleanup_jobs: 1h
```

### Best Practices

- Provide `config.example.yaml` in repo
- Use environment-specific configs (dev, staging, prod)
- Never commit actual `config.yaml` with secrets
- Support environment variable overrides
- Validate configuration on startup

---

## Database Layer

### Connection Setup

**File**: `internal/app/db.go`

```go
package app

import (
    "time"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"
    
    "yourproject/internal/config"
    "yourproject/internal/log"
)

func WithDatabase() {
    var err error
    cfg := config.C.Database
    
    A.Database, err = sqlx.Open("mysql", cfg.DSN)
    if err != nil {
        log.Logger.Fatal("failed to connect to database", zap.Error(err))
    }
    
    A.Database.SetMaxOpenConns(cfg.MaxConn)
    A.Database.SetMaxIdleConns(cfg.IdleConn)
    A.Database.SetConnMaxLifetime(cfg.Timeout)
    
    // Test connection
    if err := A.Database.Ping(); err != nil {
        log.Logger.Fatal("failed to ping database", zap.Error(err))
    }
}
```

### Migrations

Use `golang-migrate` for database migrations:

```bash
# Create migration
migrate create -ext sql -dir migrations -seq init

# Apply migrations
migrate -path migrations -database "mysql://user:pass@tcp(localhost:3306)/db" up

# Rollback
migrate -path migrations -database "mysql://user:pass@tcp(localhost:3306)/db" down 1
```

**Migration Files**:
```sql
-- migrations/000001_init.up.sql
CREATE TABLE clients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- migrations/000001_init.down.sql
DROP TABLE IF EXISTS clients;
```

### Query Patterns

```go
// Named query (INSERT/UPDATE)
query := `INSERT INTO clients (name, email) VALUES (:name, :email)`
_, err := db.NamedExecContext(ctx, query, &client)

// Get single row
query := `SELECT * FROM clients WHERE id = ?`
var client models.Client
err := db.GetContext(ctx, &client, query, id)

// Get multiple rows
query := `SELECT * FROM clients WHERE active = ?`
var clients []models.Client
err := db.SelectContext(ctx, &clients, query, true)

// IN clause with sqlx
query := `DELETE FROM clients WHERE id IN (?)`
query, args, err := sqlx.In(query, ids)
if err != nil {
    return err
}
query = db.Rebind(query)
_, err = db.ExecContext(ctx, query, args...)
```

---

## HTTP Layer

### Framework: Echo

This project uses **[Echo](https://echo.labstack.com/)** (`github.com/labstack/echo/v4`) as the HTTP framework. All HTTP handlers, middleware, and routing MUST use Echo's conventions and interfaces.

### Server Setup

**File**: `internal/http/server.go`

**Global Singleton pattern** — `NewServer()` accesses handlers via `app.A` implicitly:

```go
package http

import (
    "context"
    "net/http"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "go.uber.org/zap"

    "yourproject/internal/config"
    "yourproject/internal/http/handlers"
    "yourproject/internal/http/middlewares"
    "yourproject/internal/log"
)

type Server struct {
    echo *echo.Echo
}

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
```

**Private Dependencies pattern** — `NewServer(svc *app.Service)` creates one `*handlers.Handlers` and wires all routes from it:

```go
package http

import (
    "context"
    "net/http"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "go.uber.org/zap"

    "yourproject/internal/app"
    "yourproject/internal/config"
    "yourproject/internal/http/handlers"
    "yourproject/internal/http/middlewares"
    "yourproject/internal/log"
)

type Server struct {
    echo *echo.Echo
}

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
    api.Use(middlewares.JWTAuthentication())
    {
        api.POST("/clients", h.CreateClient)
        api.GET("/clients/:id", h.GetClient)
        api.PUT("/clients/:id", h.UpdateClient)
        api.DELETE("/clients/:id", h.DeleteClient)
    }

    return &Server{echo: e}
}
```

Both patterns share the same `Serve` and `Shutdown` methods:

```go
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

func (s *Server) Shutdown(ctx context.Context) error {
    return s.echo.Shutdown(ctx)
}
```

### Middleware Pattern

```go
// internal/http/middlewares/jwt_authentication.go
package middlewares

import (
    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt/v5"
)

func JWTAuthentication() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Authentication logic
            token := c.Request().Header.Get("Authorization")
            
            if token == "" {
                return echo.ErrUnauthorized
            }
            
            // Validate token
            claims, err := validateToken(token)
            if err != nil {
                return echo.ErrUnauthorized
            }
            
            // Set user context
            c.Set("user", claims)
            
            return next(c)
        }
    }
}
```

---

## Worker/Background Jobs

### Worker Framework

**File**: `internal/worker/worker.go`

```go
package worker

import (
    "context"
    "time"
)

type Worker interface {
    Run(handlers ...Handler)
    RunAsync(handlers ...Handler)
    Close()
}

type Handler func(ctx context.Context)

type worker struct {
    ticker *time.Ticker
    quit   chan struct{}
}

func NewWorker(interval time.Duration) Worker {
    return &worker{
        ticker: time.NewTicker(interval),
        quit:   make(chan struct{}),
    }
}

func (w *worker) Run(handlers ...Handler) {
    for {
        select {
        case <-w.ticker.C:
            for _, handler := range handlers {
                ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
                handler(ctx)
                cancel()
            }
        case <-w.quit:
            return
        }
    }
}

func (w *worker) RunAsync(handlers ...Handler) {
    go w.Run(handlers...)
}

func (w *worker) Close() {
    w.ticker.Stop()
    close(w.quit)
}
```

### Job Handler Pattern

**Global Singleton pattern**:

```go
// internal/worker/handlers/sync_databases.go
package handlers

import (
    "context"
    "go.uber.org/zap"

    "yourproject/internal/app"
    "yourproject/internal/log"
)

func SyncDatabases(ctx context.Context) {
    log.Logger.Info("starting database sync job")

    if err := app.A.Service.SyncDatabases(ctx); err != nil {
        log.Logger.Error("database sync failed", zap.Error(err))
        return
    }

    log.Logger.Info("database sync completed")
}
```

**Private Dependencies pattern** — handler is a struct receiving `*app.Service`:

```go
// internal/worker/handlers/sync_databases.go
package handlers

import (
    "context"
    "go.uber.org/zap"

    "yourproject/internal/app"
    "yourproject/internal/log"
)

type syncDatabases struct {
    svc *app.Service
}

func NewSyncDatabases(svc *app.Service) *syncDatabases {
    return &syncDatabases{svc: svc}
}

func (h *syncDatabases) Handle(ctx context.Context) {
    log.Logger.Info("starting database sync job")

    if err := h.svc.SyncDatabases(ctx); err != nil {
        log.Logger.Error("database sync failed", zap.Error(err))
        return
    }

    log.Logger.Info("database sync completed")
}
```

### Usage in Application

**Global Singleton pattern**:

```go
// cmd/start.go
func startFunc(_ *cobra.Command, _ []string) {
    app.WithGracefulShutdown()
    app.WithDatabase()
    app.WithRepository()
    app.WithService()

    if config.C.Worker.Enabled {
        syncJob := worker.NewWorker(config.C.Worker.JobsIntervals.SyncDatabases)
        syncJob.RunAsync(handlers.SyncDatabases)
        defer syncJob.Close()
    }

    app.Wait()
}
```

**Private Dependencies pattern** — see [Pattern 2: Private Dependencies](#pattern-2-private-dependencies-dependency-injection) for the full `cmd/start.go` wiring.

---

## Dependency Management

### Go Modules

```bash
# Initialize module
go mod init github.com/yourusername/projectname

# Add dependency
go get github.com/labstack/echo/v4

# Update dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

### Essential Dependencies

```go
// Web framework
github.com/labstack/echo/v4

// Logging
go.uber.org/zap

// Configuration
github.com/spf13/viper
github.com/spf13/cobra

// Database
github.com/jmoiron/sqlx
github.com/go-sql-driver/mysql

// Migrations
github.com/golang-migrate/migrate

// Metrics
github.com/prometheus/client_golang

// JWT
github.com/golang-jwt/jwt/v5

// Auto maxprocs
go.uber.org/automaxprocs
```

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External dependencies
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
    
    // Internal packages
    "yourproject/internal/app"
    "yourproject/internal/models"
)
```

---

## Build and Deployment

### Makefile

```makefile
.PHONY: build run test clean migrate-up migrate-down

APP_NAME=yourapp
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "\
    -X 'github.com/yourusername/$(APP_NAME)/internal/app.GitCommit=$(COMMIT)' \
    -X 'github.com/yourusername/$(APP_NAME)/internal/app.GitTag=$(VERSION)' \
    -X 'github.com/yourusername/$(APP_NAME)/internal/app.BuildDate=$(BUILD_TIME)'"

build:
	go build $(LDFLAGS) -o $(APP_NAME) .

run:
	go run . start --config config.yaml

test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage:
	go tool cover -html=coverage.out

clean:
	rm -f $(APP_NAME)
	rm -f coverage.out

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

docker-build:
	docker build -t $(APP_NAME):$(VERSION) .
```

### Build Information

Inject build info into the application:

```go
// internal/app/app.go
var (
    GitCommit       string
    GitRef          string
    GitTag          string
    BuildDate       string
    CompilerVersion string
)

const (
    Name    = "yourapp"
    Version = 1
)

func Banner() string {
    return fmt.Sprintf(
        "App: %s v%d\nTag: %s\nCommit: %s\nBuild: %s\n",
        Name, Version, GitTag, GitCommit, BuildDate,
    )
}
```

### Systemd Service

```ini
# deployments/systemd/yourapp.service
[Unit]
Description=Your Application Service
After=network.target mysql.service

[Service]
Type=simple
User=yourapp
Group=yourapp
WorkingDirectory=/opt/yourapp
ExecStart=/opt/yourapp/yourapp start --config /etc/yourapp/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

---

## Documentation

### Package Documentation

```go
// Package handlers provides HTTP request handlers for the application.
//
// Each handler is responsible for:
//   - Validating request input
//   - Calling appropriate service methods
//   - Formatting response output
//   - Logging errors
package handlers
```

### Function Documentation

```go
// CreateClient handles the creation of a new client.
//
// It validates the request, generates a password if not provided,
// stores the client in the database, and returns the generated links.
//
// Request body:
//   - telegram_id (string, required): Telegram user ID
//   - quota (int, required): Quota in bytes
//   - expire_days (int, required): Number of days until expiration
//   - raw_password (string, optional): Custom password
//
// Response:
//   - links ([]string): Generated connection links
//
// Returns HTTP 200 on success, 400 on validation error, 500 on server error.
func CreateClient(ctx echo.Context) error {
    // ...
}
```

### README Structure

```markdown
# Project Name

Brief description of what the project does.

## Features

- Feature 1
- Feature 2
- Feature 3

## Prerequisites

- Go 1.24+
- MySQL 8.0+
- Optional dependencies

## Installation

```bash
git clone https://github.com/yourusername/project
cd project
go mod download
```

## Configuration

Copy example config and edit:
```bash
cp config.example.yaml config.yaml
```

## Usage

Start the server:
```bash
./yourapp start --config config.yaml
```

## API Documentation

Available endpoints...

## Development

Run tests:
```bash
make test
```

## License

MIT License
```

---

## Best Practices Summary

### Code Quality

1. **Always use context.Context** as the first parameter
2. **Check errors immediately** after function calls
3. **Use interfaces** for dependencies (testability)
4. **Structured logging** with zap
5. **Named returns** only when it improves readability
6. **Table-driven tests** for comprehensive coverage

### Performance

1. **Connection pooling** for databases
2. **Graceful shutdown** for all resources
3. **Context timeouts** for operations
4. **Proper HTTP timeouts** configuration
5. **Use `automaxprocs`** for container environments

### Security

1. **Never commit secrets** to version control
2. **Use environment variables** or secure config management
3. **Validate all inputs** at API boundaries
4. **Use prepared statements** for SQL queries
5. **Implement rate limiting** and authentication

### Maintainability

1. **Small, focused functions** (< 50 lines ideal)
2. **Clear separation of concerns** between layers
3. **Consistent naming conventions** throughout
4. **Comprehensive error messages** for debugging
5. **Keep dependencies minimal** and up-to-date

---

## Checklist for New Projects

- [ ] Initialize Go module
- [ ] Set up project structure (cmd/, internal/, pkg/)
- [ ] Configure logging with zap
- [ ] Set up configuration with viper
- [ ] Create application singleton pattern
- [ ] Set up database connection and migrations
- [ ] Implement repository layer with interfaces
- [ ] Create HTTP server with middleware
- [ ] Add graceful shutdown handling
- [ ] Set up metrics and monitoring
- [ ] Create Makefile for common tasks
- [ ] Write README with setup instructions
- [ ] Add example configuration file
- [ ] Set up CI/CD pipeline
- [ ] Write unit tests for critical paths
- [ ] Document API endpoints

---

**End of Style Guide**

This guide should be treated as a living document. Update it as patterns evolve and new best practices emerge.

