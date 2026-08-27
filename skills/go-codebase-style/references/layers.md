# Layer Patterns: Handler, Repository, Service, Usecase, Models

The heart of the guide. Every layer follows the same skeleton: **interface →
private implementation struct → public constructor returning the interface**.
This is what keeps layers mockable in tests.

---

## 1. Handler Package Pattern

**Location**: `internal/http/handlers/`

**One handler per file** named after the action:
- `create_client.go`
- `update_client.go`
- `delete_client.go`
- `get_client.go`

Handlers do four things only: bind/validate input → call a **usecase** → log
errors → map sentinel errors to HTTP status codes. They never call
repositories or services directly.

### Global Singleton pattern

Handler is a plain function accessing `app.A`:

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

    result, err := app.A.Usecase.Client.Create(ctx.Request().Context(), request)
    if err != nil {
        log.Logger.Error("failed to create resource", zap.Error(err))
        return err
    }

    return ctx.JSON(http.StatusOK, CreateResourceResponse{ID: result.ID})
}
```

### Private Dependencies pattern

All handlers live as methods on a single `Handlers` struct. One constructor,
all routes covered:

**File**: `internal/http/handlers/handlers.go`

```go
package handlers

import "yourproject/internal/app"

// Handlers holds the usecase bundle and exposes all handler methods.
type Handlers struct {
    uc *app.Usecase
}

func New(uc *app.Usecase) *Handlers {
    return &Handlers{uc: uc}
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

    result, err := h.uc.Client.Create(ctx.Request().Context(), request)
    if err != nil {
        log.Logger.Error("failed to create resource", zap.Error(err))
        return err
    }

    return ctx.JSON(http.StatusOK, CreateResourceResponse{ID: result.ID})
}
```

---

## 2. Repository Package Pattern

**Location**: `internal/repository/`

Interface-driven design. Repositories own ALL data access; they translate
driver errors (`sql.ErrNoRows`) into typed sentinel errors.

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

**Key points**:
- Interface for testability
- Private implementation struct
- Public constructor returning the interface
- `context.Context` as first parameter
- Named queries with sqlx
- Map `sql.ErrNoRows` → sentinel error (see `structure-conventions.md`)

---

## 3. Service Package Pattern

**Location**: `internal/service/<name>/`

Services are **clients for external dependency services** — third-party APIs,
mail/SMS providers, message brokers, storage, etc. Each service lives in its
**own subpackage**, named after its **role** (`notifier`, `cache`, `storage`,
`mailer`) so implementations can be swapped freely. Services never touch the
database or repositories: data access is exclusively the repository layer's
job.

```go
// internal/service/notifier/notifier.go
package notifier

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// Interface definition
type Notifier interface {
    Send(ctx context.Context, chatID string, msg string) error
}

// Implementation struct (private)
type telegramNotifier struct {
    token   string
    baseURL string
    http    *http.Client
}

// Constructor returns the interface
func New(token string) Notifier {
    return &telegramNotifier{
        token:   token,
        baseURL: "https://api.telegram.org",
        http:    &http.Client{Timeout: 10 * time.Second},
    }
}

// Methods perform pure I/O against the external dependency
func (n *telegramNotifier) Send(ctx context.Context, chatID string, msg string) error {
    endpoint := n.baseURL + "/bot" + n.token + "/sendMessage"

    body, err := json.Marshal(map[string]string{"chat_id": chatID, "text": msg})
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to build request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := n.http.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send notification: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status from provider: %s", resp.Status)
    }

    return nil
}
```

**Key points**:
- One subpackage per role under `internal/service/`: `notifier/`, `cache/`,
  `storage/`, ...
- External I/O only — never import `repository`, open DB connections, or run
  SQL
- Private implementation struct; constructor returns an interface
- Wrap failures with context (`fmt.Errorf("...: %w", err)`); callers decide
  retry/log policy
- Usecases consume these clients together with repositories

---

## 4. Usecase Package Pattern

**Location**: `internal/usecase/`

The usecase package contains the **big logic**: multi-step flows, cross-domain
orchestration, policy decisions. HTTP handlers and worker/task handlers call
usecases — not services or repositories directly. A single usecase flow can
combine **repositories** (data access) and **service clients** (external
dependencies) at the same time.

```go
package usecase

import (
    "context"
    "fmt"

    "go.uber.org/zap"

    "yourproject/internal/log"
    "yourproject/internal/models"
    "yourproject/internal/repository"
    "yourproject/internal/service/notifier"
)

// Request/response types owned by the usecase
type CreateClientRequest struct {
    Name       string `json:"name"`
    Email      string `json:"email"`
    TelegramID string `json:"telegram_id"`
}

// Interface definition
type ClientUsecase interface {
    Create(ctx context.Context, req CreateClientRequest) (*models.Client, error)
}

// Implementation struct (private)
type clientUsecase struct {
    repo     repository.Client
    notifier notifier.Notifier
}

// Constructor
func NewClientUsecase(repo repository.Client, notifier notifier.Notifier) ClientUsecase {
    return &clientUsecase{
        repo:     repo,
        notifier: notifier,
    }
}

// Big flow: persist via repository + call external service clients in one place
func (u *clientUsecase) Create(ctx context.Context, req CreateClientRequest) (*models.Client, error) {
    client := models.Client{Name: req.Name, Email: req.Email}
    if err := u.repo.Create(ctx, client); err != nil {
        return nil, fmt.Errorf("failed to create client: %w", err)
    }

    // External dependency used right alongside data access
    if err := u.notifier.Send(ctx, req.TelegramID, "Welcome!"); err != nil {
        log.Logger.Error("failed to send welcome notification", zap.Error(err))
        // Non-fatal: continue the flow
    }

    // ...compose more repositories/service clients as the flow grows...

    return &client, nil
}
```

**Key points**:
- Dependency direction: handlers → **usecase** → repository / service — never
  upward
- One file per usecase: `client.go`, `auth.go`
- Keep request/response DTO types next to the usecase that owns them
- Usecases may call repositories AND external service clients in the same flow
- Usecases decide fatality: log-and-continue for optional side effects
  (notifications, cache refresh), propagate errors for required ones

---

## 5. Models Package Pattern

**Location**: `internal/models/`

```go
package models

import "time"

type Client struct {
    ID        uint32     `db:"id" json:"id"`
    Name      string     `db:"name" json:"name"`
    Email     string     `db:"email" json:"email"`
    CreatedAt time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}
```

**Guidelines**:
- Struct tags for both `db` and `json`
- Pointers for nullable fields
- Keep models simple (no business logic)
