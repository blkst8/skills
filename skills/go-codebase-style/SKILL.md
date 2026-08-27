---
name: go-codebase-style
description: >
  Apply the blkst8 Go codebase style guide when generating, scaffolding, or
  reviewing Go code. ALWAYS use this skill when the user asks to create a new Go
  project or service, scaffold any Go code, add a handler / repository / service
  / usecase / model / worker / migration, set up HTTP routing with Echo,
  configure logging (zap) or database (sqlx / MySQL), wire dependencies, or
  structure a Go application — even if they don't explicitly mention the style
  guide. Also trigger when the user says "follow the style guide", "follow our
  Go conventions", "use our patterns", "how should I structure X in Go", or asks
  about private DI vs global singleton in this codebase.
---

# Go Codebase Style Guide (blkst8, v1.1)

Six focused reference files back this skill. Don't load them all — pick the one
that matches the task, using this map:

| Reference | Load it when... |
|---|---|
| `references/structure-conventions.md` | Scaffolding or reorganizing a project; naming questions (files, packages, receivers); import ordering; error-handling patterns (sentinel errors, HTTP error mapping) |
| `references/wiring-patterns.md` | Creating/modifying `internal/app/` or `cmd/start.go`; wiring dependencies; deciding between Global Singleton and Private DI; graceful shutdown |
| `references/layers.md` | Writing or reviewing any handler, repository, service, usecase, or model code — the heart of the guide |
| `references/http-workers.md` | Echo server setup, routes, middleware; background jobs (ticker workers or worker pool) |
| `references/config-logging-db.md` | Viper config structs / `config.example.yaml`; zap logging setup or usage; DB connections, migrations, sqlx query patterns |
| `references/build-deploy-docs.md` | Makefile, ldflags/build info, systemd units; go.mod dependencies; godoc conventions; README template; project checklist |

If anything here feels ambiguous, the reference files are the source of truth —
prefer them over guessing. They contain complete copy-paste-ready examples.

---

## Step 0 — Clarify dependency wiring (ask once, remember for the session)

Two wiring styles are supported and they shape every file you write. If the user
hasn't said which one, ask before writing code:

> "Should I use **global singleton** (`app.A.*`) or **private dependency
> injection** (structs wired in `cmd/start.go`)?"

| | Global Singleton | Private DI |
|---|---|---|
| Access | `app.A.*` package-level var | Injected via constructors |
| Testability | Low | High |
| Best for | Small scripts / simple services | Production services, multi-instance, testable code |

Once chosen, apply it **consistently everywhere** in the session. Never mix
patterns in one codebase.

---

## Non-negotiable rules (apply to ALL code)

1. **Layer chain**: HTTP handlers and job/task handlers call **usecases**.
   Usecases contain the big business logic and combine **repositories** (data
   access) with **services** (clients for external dependency services).
   Dependencies never point upward.
2. **Services are external-I/O only**: one package per role under
   `internal/service/` (`notifier/`, `cache/`, `storage/`, `mailer/`). They
   never import repositories, open DB connections, or run SQL. Data access is
   exclusively the repository layer's job.
3. **`context.Context` is the first parameter** of every I/O function.
4. **Errors**: check immediately after each call; wrap with
   `fmt.Errorf("failed to X: %w", err)` when adding context; define sentinel
   errors in repositories (`var ErrXxx = errors.New(...)`); handlers map them
   to `echo.NewHTTPError(status, msg)` via `errors.Is`.
5. **Logging**: only `log.Logger` from `internal/log` with structured zap
   fields (`zap.Error(err)`, `zap.String("k", v)`). Never `fmt.Println` or the
   stdlib `log` package.
6. **Models**: dumb structs with both `db:` and `json:` tags; pointer types for
   nullable fields; no business logic.
7. **Database**: always sqlx; named queries for INSERT/UPDATE
   (`NamedExecContext`), `GetContext`/`SelectContext` for reads.
8. **Imports**: stdlib → external → internal, blank line between groups.
9. **Naming**: snake_case files named after their primary action
   (`create_client.go`); lowercase single-word packages; receivers are short
   1-2 letter abbreviations, never `this`/`self`.
10. **Interfaces for dependencies**: every repository/service/usecase is an
    interface + private implementation struct + public constructor returning
    the interface. This is what makes the code testable.

---

## Project layout (compact)

```
project/
├── cmd/                   # Cobra commands: root.go, start.go, migrate.go
├── internal/
│   ├── app/               # Wiring: app.go / db.go / repository.go / service.go / usecase.go
│   ├── config/            # Viper config structs (+ builtin.go defaults)
│   ├── http/              # server.go + handlers/ (one file per action) + middlewares/
│   ├── log/               # zap setup (log.go)
│   ├── metrics/           # business.go + monitoring.go
│   ├── models/            # plain structs, db+json tags
│   ├── repository/        # sqlx data access, interface-driven
│   ├── service/           # EXTERNAL clients only, one package per role
│   │   ├── notifier/      #   interface + private impl + constructor
│   │   └── cache/
│   ├── usecase/           # big business logic; one file per domain (client.go)
│   ├── worker/            # ticker-based periodic jobs + handlers/
│   └── workers/           # goroutine pool: pool.go + tasks.go
├── pkg/                   # public reusable libraries
├── migrations/            # YYYYMMDDHHMMSS_name.{up,down}.sql
├── docs/                  # features/ + architecture/
├── deployments/           # systemd/ service files
├── main.go                # entry point
├── Makefile
├── config.example.yaml
└── README.md
```

---

## Layer responsibilities (who talks to whom)

- **Handlers** (`internal/http/handlers/`): bind/validate request → call
  `usecase` → log error → map sentinel errors to HTTP status → respond.
  Never touch repositories or services directly.
- **Usecases** (`internal/usecase/`): the big flows. A single flow can call
  repositories AND service clients together; decides what is fatal vs
  log-and-continue. One file per domain.
- **Repositories** (`internal/repository/`): sqlx queries, rows → models,
  translate `sql.ErrNoRows` into typed sentinel errors.
- **Services** (`internal/service/<role>/`): pure external I/O (HTTP calls,
  brokers, storage). Wrap failures with `%w`; callers decide retry/log policy.
- **Job/task handlers** (`internal/worker/handlers/`, `internal/workers/`):
  same rule as HTTP handlers — they call usecases only.

Full copy-paste examples for every layer: `references/layers.md`.

---

## Wiring quick reference

**Private DI** — `internal/app/` holds typed bundles, no global state:

- `db.go`: `WithDatabase() *sqlx.DB`
- `repository.go`: `Repository` struct + `WithRepository(db) *Repository`
- `service.go`: `Service` struct (external clients only) + `WithServices() *Service`
- `usecase.go`: `Usecase` struct + `WithUsecases(repo, svc) *Usecase`
- `cmd/start.go`: chain them; pass `*app.Usecase` into the HTTP server, ticker
  workers, and worker pool.

**Global singleton** — `app.A` holds everything
(`Database`, `Repository`, `Service`, `Usecase`); `WithDatabase()`-style
builders mutate it; handlers use `app.A.Usecase.Client.Create(...)`.

Full code for both patterns including `cmd/start.go` with graceful shutdown:
`references/wiring-patterns.md`.
