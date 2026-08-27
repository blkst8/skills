# HTTP Layer & Background Jobs

Covers: Echo server setup (both wiring patterns), middleware, the ticker-based
worker framework, and the worker pool.

---

## HTTP Framework: Echo

This codebase uses **[Echo](https://echo.labstack.com/)**
(`github.com/labstack/echo/v4`) as the HTTP framework. All HTTP handlers,
middleware, and routing use Echo's conventions and interfaces.

## Server Setup

**File**: `internal/http/server.go`

### Global Singleton pattern

`NewServer()` accesses handlers via `app.A` implicitly:

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

### Private Dependencies pattern

`NewServer(uc *app.Usecase)` creates one `*handlers.Handlers` and wires all
routes from it:

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
        api.POST("/clients", h.CreateClient)
        api.GET("/clients/:id", h.GetClient)
        api.PUT("/clients/:id", h.UpdateClient)
        api.DELETE("/clients/:id", h.DeleteClient)
    }

    return &Server{echo: e}
}
```

### Serve and Shutdown (both patterns)

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

The request-logging middleware (`middlewares.ZapLogger`) implementation is in
`references/config-logging-db.md` under Logging.

---

## Middleware Pattern

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

## Background Jobs: Two Mechanisms

Two background-execution mechanisms are provided — choose per project and
apply consistently:

- **`internal/worker`** (ticker): runs handlers on a fixed interval — best for
  periodic jobs
- **`internal/workers`** (pool): fixed goroutine pool consuming a bounded task
  queue with timeout/retry — best for concurrent, on-demand work

Both call into the **usecase layer**, never repositories or services directly.

---

## Worker Framework (ticker)

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

    if err := app.A.Usecase.Client.SyncDatabases(ctx); err != nil {
        log.Logger.Error("database sync failed", zap.Error(err))
        return
    }

    log.Logger.Info("database sync completed")
}
```

**Private Dependencies pattern** — handler is a struct receiving
`*app.Usecase`:

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
    uc *app.Usecase
}

func NewSyncDatabases(uc *app.Usecase) *syncDatabases {
    return &syncDatabases{uc: uc}
}

func (h *syncDatabases) Handle(ctx context.Context) {
    log.Logger.Info("starting database sync job")

    if err := h.uc.Client.SyncDatabases(ctx); err != nil {
        log.Logger.Error("database sync failed", zap.Error(err))
        return
    }

    log.Logger.Info("database sync completed")
}
```

### Usage in Application (ticker)

**Global Singleton pattern**:

```go
// cmd/start.go
func startFunc(_ *cobra.Command, _ []string) {
    app.WithGracefulShutdown()
    app.WithDatabase()
    app.WithRepository()
    app.WithService()
    app.WithUsecase()

    // Interval-based workers
    if config.C.Worker.Enabled {
        syncJob := worker.NewWorker(config.C.Worker.JobsIntervals.SyncDatabases)
        syncJob.RunAsync(handlers.SyncDatabases)
        defer syncJob.Close()
    }

    app.Wait()
}
```

**Private Dependencies pattern** — see `references/wiring-patterns.md` for the
full `cmd/start.go` wiring.

---

## Worker Pool (concurrent, on-demand)

For concurrent, on-demand task execution use `internal/workers`. The pool runs
a fixed number of goroutines that consume tasks from a bounded queue, applying
a per-task timeout and retry count taken from the `config` package.

**File**: `internal/workers/pool.go`

```go
package workers

import (
    "context"
    "errors"
    "sync"

    "go.uber.org/zap"

    "yourproject/internal/config"
    "yourproject/internal/log"
)

// Task is a unit of work executed by the pool.
type Task interface {
    Name() string
    Execute(ctx context.Context) error
}

var ErrQueueFull = errors.New("workers queue is full")

// Pool is a fixed-size worker pool with a bounded queue.
type Pool struct {
    cfg   config.Workers
    tasks chan Task
    wg    sync.WaitGroup
}

func NewPool(cfg config.Workers) *Pool {
    return &Pool{
        cfg:   cfg,
        tasks: make(chan Task, cfg.QueueSize),
    }
}

// Start launches the configured number of worker goroutines.
func (p *Pool) Start() {
    for i := 0; i < p.cfg.Count; i++ {
        p.wg.Add(1)
        go p.loop(i)
    }
}

func (p *Pool) loop(id int) {
    defer p.wg.Done()

    for task := range p.tasks {
        for attempt := 0; attempt <= p.cfg.Retries; attempt++ {
            ctx, cancel := context.WithTimeout(context.Background(), p.cfg.Timeout)
            err := task.Execute(ctx)
            cancel()
            if err == nil {
                break
            }
            log.Logger.Error("task failed",
                zap.Int("worker_id", id),
                zap.String("task", task.Name()),
                zap.Int("attempt", attempt+1),
                zap.Error(err),
            )
        }
    }
}

// Submit enqueues a task without blocking.
// Returns ErrQueueFull when the queue has no capacity left.
// Must not be called after Stop.
func (p *Pool) Submit(task Task) error {
    select {
    case p.tasks <- task:
        return nil
    default:
        log.Logger.Error("submit failed", zap.String("task", task.Name()), zap.Error(ErrQueueFull))
        return ErrQueueFull
    }
}

// Stop closes the queue and waits until running/pending tasks complete.
func (p *Pool) Stop() {
    close(p.tasks)
    p.wg.Wait()
}
```

**File**: `internal/workers/tasks.go` — one file per task; constructors
receive the dependencies (typically the usecase bundle):

```go
package workers

import (
    "context"
    "go.uber.org/zap"

    "yourproject/internal/app"
    "yourproject/internal/log"
)

type syncDatabasesTask struct {
    uc *app.Usecase
}

func NewSyncDatabasesTask(uc *app.Usecase) *syncDatabasesTask {
    return &syncDatabasesTask{uc: uc}
}

func (t *syncDatabasesTask) Name() string { return "sync_databases" }

func (t *syncDatabasesTask) Execute(ctx context.Context) error {
    log.Logger.Info("running database sync task")
    return t.uc.Client.SyncDatabases(ctx)
}
```

**Global Singleton pattern** — the task reads `app.A` directly instead:

```go
func (t *syncDatabasesTask) Execute(ctx context.Context) error {
    return app.A.Usecase.Client.SyncDatabases(ctx)
}
```

**Key points**:
- All sizing knobs come from `config.C.Workers` (`enabled`, `count`,
  `queue_size`, `timeout`, `retries`)
- Tasks implement the `Task` interface — one small type per job, named by its
  action
- Tasks call **usecases**, never repositories or services directly
- `Submit` is non-blocking: a full queue returns `ErrQueueFull` so callers can
  decide to drop or back off
- `Stop()` drains the queue gracefully; pair it with a context timeout in
  `startFunc`
