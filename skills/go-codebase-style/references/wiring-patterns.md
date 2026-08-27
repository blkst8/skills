# Dependency Wiring Patterns

The **Database**, **Repository**, **Service**, **Usecase**, and **Workers**
layers support two wiring strategies. **Choose one per project** and apply it
consistently — never mix patterns in one codebase.

| | Global Singleton | Private Dependencies |
|---|---|---|
| Access | `app.A.*` global | Injected via constructors |
| Testability | Low | High |
| Best for | Simple scripts / small services | Production services, multi-instance, testable code |

---

## Pattern 1: Global Singleton

**Purpose**: centralized access to application-wide dependencies via a
package-level variable. Fast to wire; accept the testability trade-off.

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
    Usecase      *Usecase

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
func WithUsecase()          { /* ... */ }
func Wait()                 { /* ... */ }
```

**Usage in `cmd/start.go`**:

```go
func startFunc(_ *cobra.Command, _ []string) {
    app.WithGracefulShutdown()
    app.WithDatabase()
    app.WithRepository()
    app.WithService()
    app.WithUsecase()

    srv := httpserver.NewServer()
    srv.Serve()

    app.Wait()
}
```

**Handlers access dependencies via the global** (always through the usecase
layer):

```go
func CreateClient(ctx echo.Context) error {
    result, err := app.A.Usecase.Client.Create(ctx.Request().Context(), req)
    // ...
}
```

---

## Pattern 2: Private Dependencies (Dependency Injection)

**Purpose**: each layer receives its dependencies explicitly. No global state.
Use for production services requiring testability and clean boundaries.

The `app` package exposes typed structs (`Repository`, `Service`, `Usecase`)
and constructor functions (`WithDatabase`, `WithRepository`, `WithServices`,
`WithUsecases`) that return concrete instances. There is no `application`
singleton struct.

### `internal/app/db.go`

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

### `internal/app/repository.go`

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

### `internal/app/service.go`

Services are clients for **external dependency services** only — they never
receive the database or repositories. Data access is exclusively the
repository layer's job.

```go
package app

import (
    "yourproject/internal/service/notifier"
)

// Service holds all clients for external dependencies (third-party APIs,
// mail/SMS providers, message brokers, storage, ...).
type Service struct {
    Notifier notifier.Notifier
    // Add other external service clients here (cache, storage, ...)
}

// WithServices constructs all external service clients and returns the bundle.
func WithServices() *Service {
    return &Service{
        Notifier: notifier.New("bot-token-from-config"),
    }
}
```

### `internal/app/usecase.go`

Usecases combine **both**: repositories for data access AND services for
external dependencies.

```go
package app

import (
    "yourproject/internal/usecase"
)

// Usecase holds all usecase instances. Usecases carry the big business logic;
// they are the only layer HTTP handlers and worker/task handlers talk to.
type Usecase struct {
    Client usecase.ClientUsecase
    // Add other usecases here
}

// WithUsecases constructs all usecases and returns the bundle.
func WithUsecases(repo *Repository, svc *Service) *Usecase {
    return &Usecase{
        Client: usecase.NewClientUsecase(repo.Client, svc.Notifier),
    }
}
```

### `cmd/start.go` — wiring everything together

```go
func startFunc(_ *cobra.Command, _ []string) {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    db   := app.WithDatabase()
    defer db.Close()

    repo := app.WithRepository(db)
    svc  := app.WithServices()
    uc   := app.WithUsecases(repo, svc)

    // Start interval-based workers (ticker jobs)
    if config.C.Worker.Enabled {
        syncJob := worker.NewWorker(config.C.Worker.JobsIntervals.SyncDatabases)
        syncJob.RunAsync(workerhandlers.NewSyncDatabases(uc).Handle)
        defer syncJob.Close()
    }

    // Start the worker pool
    if config.C.Workers.Enabled {
        pool := workers.NewPool(config.C.Workers)
        pool.Start()
        defer pool.Stop()
        if err := pool.Submit(workers.NewSyncDatabasesTask(uc)); err != nil {
            log.Logger.Error("failed to submit task", zap.Error(err))
        }
    }

    // Start HTTP server
    srv := httpserver.NewServer(uc)
    srv.Serve()
    defer srv.Shutdown(context.Background())

    <-ctx.Done()
}
```

---

## Which pattern when

- Default recommendation for anything production-bound: **Private DI** — the
  `cmd/start.go` chain (`db → repo → svc → uc → server/workers`) makes every
  dependency visible and every layer mockable in tests.
- **Global Singleton** is fine for small utilities, internal tools, and
  scripts where the ceremony of DI isn't worth it.

Build info variables (`GitCommit`, `GitTag`, `BuildDate`) and the `Banner()`
helper live in `app.go` too — see `references/build-deploy-docs.md`.
