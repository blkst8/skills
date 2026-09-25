---
name: go-codebase-style
description: >
  Apply the blkst8 Go codebase style guide when generating, scaffolding, or
  reviewing Go code. ALWAYS use this skill when the user asks to create a new Go
  project or service, scaffold any Go code, add a handler / repository / service
  / model / worker / migration, set up HTTP routing with Echo, configure logging
  (zap) or database (sqlx / MySQL), wire dependencies, or structure a Go
  application — even if they don't explicitly mention the style guide. Also
  trigger when the user says "follow the style guide", "follow our Go
  conventions", "use our patterns", "how should I structure X in Go", or asks
  about private DI vs singleton in this codebase.
---

# Go Codebase Style Guide Skill

Full reference spec: `references/style-guide.md` (1500+ lines) — load it when
you need exhaustive code examples, complete file templates, or anything not
covered by the quick-reference sections below.

**Table of contents for `references/style-guide.md`:**

| Section | What's there |
|---|---|
| Project Structure | Full annotated directory tree |
| Architecture Patterns | Layered arch diagram; full Global Singleton + Private DI code (db.go, repository.go, service.go, cmd/start.go) |
| Code Organization | Full handler, repository, and model code examples |
| Naming Conventions | Files, packages, variables, functions, receivers |
| Error Handling | Standard pattern, repository errors, HTTP error mapping |
| Logging | Zap setup (internal/log/log.go), usage patterns, request-logging middleware |
| Configuration Management | Full Viper config structs, YAML example, best practices |
| Database Layer | Connection setup, migration commands, query patterns (named, get, select, IN clause) |
| HTTP Layer | Full server.go template, middleware pattern |
| Worker / Background Jobs | Worker framework, job handler pattern, usage in start.go |
| Dependency Management | go mod commands, full list of essential dependencies, import ordering |
| Build and Deployment | Makefile template, build info injection, systemd service file |
| Documentation | Package/function doc conventions, README structure |

---

## Step 0 — Clarify dependency wiring (ask once, remember for the session)

Before writing any code, determine which pattern applies. If the user hasn't
said, ask:

> "Should I use **global singleton** (`app.A.*`) or **private dependency
> injection** (structs wired in `cmd/start.go`)?"

| | Global Singleton | Private DI |
|---|---|---|
| Access | `app.A.*` package-level var | Injected via constructors |
| `app` struct | Yes (`application` struct + `var A`) | No struct — just typed bundles |
| Best for | Small / simple services | Production, testable, multi-instance |

Once chosen, apply it consistently everywhere in the session.

---

## Project Layout (always apply this)

```
project/
├── cmd/                   # Cobra commands: root.go, start.go, migrate.go
├── internal/
│   ├── app/               # app.go, db.go, repository.go, service.go
│   ├── config/            # config.go, builtin.go
│   ├── http/
│   │   ├── server.go
│   │   ├── handlers/      # handlers.go (Handlers struct) + one file per action
│   │   └── middlewares/
│   ├── log/               # log.go (zap setup)
│   ├── models/            # plain structs with db/json tags
│   ├── repository/        # interfaces + implementations
│   └── worker/
│       ├── worker.go
│       └── handlers/      # job handler structs
├── pkg/                   # Public reusable packages
├── migrations/
├── doc/
├── deployments/
├── main.go
├── Makefile
└── config.example.yaml
```

Key principles:
- `internal/` — code that must not be imported by other projects
- `pkg/` — reusable libraries that can be imported externally
- `cmd/` — each file = one Cobra command

---

## Core Patterns (quick reference)

### Global Singleton — `internal/app/`

`app.go` declares the `application` struct and `var A *application`; `init()`
sets it. `WithDatabase()`, `WithRepository()`, `WithService()`,
`WithGracefulShutdown()`, and `Wait()` mutate `A` in place. Handlers call
`app.A.Service.X` directly.

### Private DI — `internal/app/`

No `application` struct, no global. Three files, three typed structs:

- **`db.go`** — `WithDatabase() *sqlx.DB`
- **`repository.go`** — `Repository` struct + `WithRepository(db) *Repository`
  — contains one field per repository interface
- **`service.go`** — `Service` struct + `WithServices(db, repo) *Service`
  — contains all service instances **and** shared infra clients (Redis, tracer,
  etc.)

`cmd/start.go` calls them in order, passing each result to the next, and
threads the resulting `*Service` into the HTTP server and worker constructors.

### HTTP Handlers

**Both patterns**: one file per handler action (`create_client.go`,
`get_client.go`, …). Request/Response types are local to each file.

**Global**: plain `func CreateClient(ctx echo.Context) error` using `app.A.Service`.

**Private DI**: a single `Handlers` struct in `handlers/handlers.go`:
```go
type Handlers struct { svc *app.Service }
func New(svc *app.Service) *Handlers { return &Handlers{svc: svc} }
```
Each action file adds a method: `func (h *Handlers) CreateClient(ctx echo.Context) error`.
`NewServer(svc)` calls `h := handlers.New(svc)` once, then wires `h.CreateClient`, etc.

### Repository

Always interface-driven:
```go
type Client interface {
    Create(ctx context.Context, client models.Client) error
    Get(ctx context.Context, id string) (*models.Client, error)
}
type client struct{ db *sqlx.DB }
func NewClientRepository(db *sqlx.DB) Client { return &client{db: db} }
```
Use `context.Context` as first param everywhere. Use `sqlx` named queries.
Map `sql.ErrNoRows` to a typed sentinel error (`var ErrClientNotFound = errors.New(...)`).

### Workers (background jobs)

**Private DI**: job handler is a struct with the service injected:
```go
type syncDatabases struct{ svc *app.Service }
func NewSyncDatabases(svc *app.Service) *syncDatabases { ... }
func (h *syncDatabases) Handle(ctx context.Context) { ... }
```
**Global**: plain `func SyncDatabases(ctx context.Context)` using `app.A.Service`.

---

## Naming & Style Rules

- Files: snake_case (`create_client.go`, `jwt_authentication.go`)
- Packages: lowercase single word, plural for collections (`handlers`, `middlewares`)
- Exported: PascalCase; unexported: camelCase
- Receiver names: short 1-2 letter abbreviation, never `this`/`self`
- Import order: stdlib → external → internal (blank line between groups)
- Constants: PascalCase preferred; `SCREAMING_SNAKE_CASE` acceptable

---

## Logging (zap)

Always use `log.Logger` from `internal/log`. Structured fields only:
```go
log.Logger.Error("failed to create client", zap.Error(err))
log.Logger.Info("server started", zap.String("address", cfg.Listen))
```
Never use `fmt.Println` or the standard `log` package in application code.

---

## Error Handling Rules

1. Check errors immediately — no deferred checks.
2. Use `fmt.Errorf("...context...: %w", err)` to wrap with context.
3. Define sentinel errors in the repository layer (`var ErrXxx = errors.New(...)`).
4. In HTTP handlers: map sentinel errors to `echo.NewHTTPError(statusCode, msg)`.

---

## Configuration

Use Viper. `config.go` defines a `Config` struct with nested sub-structs for
`Logger`, `HTTPServer`, `Database`, `Worker`. `var C *Config` is the singleton.
Always provide `config.example.yaml`. Never commit real secrets.

---

## When to read `references/style-guide.md`

The quick-reference above is sufficient for most tasks. Load the full spec when:

- You need a **complete, copy-paste-ready file** (e.g. full `server.go`, `app.go`, `Makefile`)
- The user asks about **database query patterns** (named exec, IN clause, migrations)
- You need the **full list of Go module dependencies** with exact import paths
- The user asks about **build flags, ldflags, systemd deployment**, or CI/CD setup
- You need to write **package-level or function-level godoc** comments
- Anything feels ambiguous — the spec is the source of truth; prefer it over guessing
