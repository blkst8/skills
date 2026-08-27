# Structure & Conventions

Covers: project layout, layer boundaries, naming conventions, import
organization, and error-handling patterns.

---

## Standard Directory Layout

```
project/
├── cmd/                        # Command-line interface commands
│   ├── root.go                # Root command (cobra)
│   ├── start.go               # Start server command
│   ├── migrate.go             # Database migration command
│   └── jwt.go                 # Utility commands
├── internal/                   # Private application code
│   ├── app/                   # Global application setup / wiring
│   │   ├── app.go             # Application singleton (or DI bundles)
│   │   ├── db.go              # Database initialization
│   │   ├── repository.go      # Repository initialization
│   │   ├── service.go         # Service initialization
│   │   └── usecase.go         # Usecase initialization
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
│   ├── service/               # Clients for external dependency services — one package per role
│   │   ├── notifier/          # e.g. external API client (Telegram, ...)
│   │   │   └── notifier.go    # Interface + private implementation + constructor
│   │   └── cache/             # e.g. infra-tech client wrapper
│   │       └── cache.go
│   ├── usecase/               # Big business logic and flows (single package)
│   │   └── client.go          # One file per usecase/domain
│   ├── worker/                # Interval-based background jobs (ticker)
│   │   ├── worker.go          # Worker framework
│   │   └── handlers/          # Job handlers
│   └── workers/               # Worker pool (concurrent task execution)
│       ├── pool.go            # Pool implementation
│       └── tasks.go           # Task definitions
├── pkg/                       # Public library code
│   └── client/                # Reusable client packages
├── migrations/                # Database migrations
│   ├── YYYYMMDDHHMMSS_name.up.sql
│   └── YYYYMMDDHHMMSS_name.down.sql
├── docs/                       # Project documentation
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

1. **`internal/` for private code**: code that should not be imported by other
   projects
2. **`pkg/` for public libraries**: reusable packages that can be imported
3. **`cmd/` for executables**: each file is a Cobra command
4. **Separation of concerns**: clear boundaries between layers
5. **Layer boundaries**: HTTP handlers and job/task handlers call
   **usecases**; usecases contain the big business logic and combine
   **repositories** (data access) with **services** (clients for external
   dependency services) — dependencies never point upward

---

## Layered Architecture

```
┌─────────────────────────────────────┐
│         HTTP/CLI Layer              │  (cmd/, internal/http/, internal/worker/handlers/, internal/workers/tasks.go)
├─────────────────────────────────────┤
│         Usecase Layer               │  (internal/usecase/) — big business logic and flows
├─────────────────────────────────────┤
│         Service Layer               │  (internal/service/<name>/) — clients for external dependency services
├─────────────────────────────────────┤
│         Repository Layer            │  (internal/repository/)
├─────────────────────────────────────┤
│         Models Layer                │  (internal/models/)
├─────────────────────────────────────┤
│         Database Layer              │  (internal/app/db.go)
└─────────────────────────────────────┘
```

---

## Naming Conventions

### Files

- **Snake case**: `create_client.go`, `jwt_authentication.go`
- **Descriptive names**: name files after their primary function
- **Test files**: `*_test.go`

### Packages

- **Lowercase, single word**: `handlers`, `models`, `repository`
- **Plural for collections**: `handlers`, `middlewares`
- **Avoid generic names**: use `httpserver` instead of `server`

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

// Interfaces: noun or adjective — no -er suffix preferred
type Client interface {}  // Good
type Reader interface {}  // Acceptable
```

### Receivers

- **Short and consistent**: 1-2 letter abbreviations
- **Never `this` or `self`**

```go
type Client struct {}

func (c *client) Get() {}   // Good
func (cl *client) Get() {}  // OK if 'c' conflicts
func (this *client) Get() {} // Bad
```

---

## Import Organization

Three groups, blank line between each: standard library, external, internal.

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

// Wrap errors with context (when adding value for the caller)
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}
```

### Repository Error Handling

Map low-level driver errors to typed sentinel errors so upper layers don't
need to know about `database/sql` internals:

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

Handlers translate sentinel errors into HTTP status codes with
`errors.Is` + `echo.NewHTTPError`:

**Global Singleton pattern**:

```go
func Handler(ctx echo.Context) error {
    result, err := app.A.Usecase.Client.DoSomething()
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

**Private DI pattern** — same logic inside a handler struct method:

```go
func (h *Handlers) Handle(ctx echo.Context) error {
    result, err := h.uc.Client.DoSomething()
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
