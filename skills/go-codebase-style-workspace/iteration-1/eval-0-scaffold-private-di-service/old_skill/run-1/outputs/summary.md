# invoice-service scaffold — summary

Scaffolded a complete Go service (`github.com/blkst8/invoice-service`) following the
`go-codebase-style` skill, using the **private dependency injection** pattern (the user's
explicit choice from Step 0 — no `app.A` global, no `application` struct).

## What was built

**Dependency wiring (private DI, per `internal/app/`):**
- `internal/app/db.go` — `WithDatabase() *sqlx.DB` (MySQL via go-sql-driver, pool settings, ping)
- `internal/app/repository.go` — `Repository` struct + `WithRepository(db)` (one field per repo interface)
- `internal/app/service.go` — `Service` struct + `WithServices(db, repo)` (service instances + shared infra)
- `internal/app/app.go` — build-info vars (GitCommit/GitTag/BuildDate) injected by the Makefile ldflags + `Banner()`
- `cmd/start.go` — wires everything in order: signal context → `WithDatabase` → `WithRepository` → `WithServices` → reconciliation worker (5m interval, gated on `worker.enabled`) → `httpserver.NewServer(svc).Serve()`; deferred close/shutdown run after `<-ctx.Done()` for graceful shutdown

**HTTP layer (Echo, private-DI handler struct):**
- `internal/http/server.go` — `NewServer(svc *app.Service)` creates one `handlers.New(svc)`, registers Recover/RequestID/ZapLogger/CORS middlewares, `/healthz`, `/metrics` (Prometheus), and `/api/v1` invoice routes
- `internal/http/handlers/` — `handlers.go` (`Handlers` struct) plus one file per action: `create_invoice.go`, `get_invoice.go`, `list_invoices.go`, plus package-level `healthz.go`, `metrics.go`
- `internal/http/middlewares/` — `logger.go` (ZapLogger request logging) and `jwt_authentication.go` (HS256 Bearer JWT, applied only when `http_server.jwt_secret` is set so the scaffold runs out of the box)

**Data layer:**
- `internal/models/invoice.go` — plain struct with `db`/`json` tags, nullable `updated_at`
- `internal/repository/invoice.go` — interface + private sqlx implementation; named queries, `GetContext`/`SelectContext`, `sql.ErrNoRows` → sentinel `ErrInvoiceNotFound`, `MarkOverdue` for reconciliation
- `internal/services/invoice.go` — `InvoiceService` interface; validation (sentinel `ErrInvalidInvoice` → 400), error wrapping with `%w`, pagination clamping, invoice-number generation
- `migrations/000001_init.{up,down}.sql` — `invoices` table with status/due_at index

**Worker:**
- `internal/worker/worker.go` — ticker-based framework (`Run`/`RunAsync`/`Close`, 2-minute per-run timeout)
- `internal/worker/handlers/reconcile_invoices.go` — struct handler receiving `*app.Service`, logs reconciled count

**Infra/config/build:**
- `cmd/root.go`, `cmd/start.go`, `cmd/migrate.go` (one Cobra command per file); `migrate` applies migrations programmatically via golang-migrate
- `internal/config/config.go` + `builtin.go` — Viper YAML config (`Logger`, `HTTPServer`, `Database`, `Worker.JobsIntervals.ReconcileInvoices`) with defaults; `config.example.yaml`
- `internal/log/log.go` — zap production logger + `SetLevel` driven by config
- `main.go` (with `automaxprocs`), `go.mod`/`go.sum` (resolved with `go mod tidy`), `Makefile` (build/run/test/migrate/lint targets, ldflags build-info), `README.md`, `deployments/systemd/invoice-service.service`

## Decisions worth noting

1. **Viper tags**: added `mapstructure` tags alongside the guide's `yaml` tags — `viper.Unmarshal` only honors `mapstructure`, so snake_case keys (`read_timeout`, `jobs_intervals`) would silently fail to bind with `yaml` tags alone.
2. **JWT optional**: the guide's server template includes `JWTAuthentication()`, which needs a secret; it's wired conditionally on `http_server.jwt_secret` so an empty secret keeps `/api/v1` open for local development.
3. **`cmd/migrate.go`**: implemented `migrate up` in-process via golang-migrate (the layout lists the command; the guide lists golang-migrate as an essential dep), with a `Database.MigrateDSN()` helper that prefixes the app DSN with `mysql://`.
4. **Repo `Create` returns `(int64, error)`** (LastInsertId) instead of the guide's bare `error` so the service/handler can return the generated invoice id and number in the 201 response.
5. **Service keeps the `db` handle** (per the guide's `NewClientService(db, repo.Client)` pattern) for future transactional methods; reconciliation itself is a single indexed UPDATE returning `RowsAffected`.
6. Empty `pkg/` and `doc/` directories from the layout were omitted (nothing to put in them yet); `deployments/` and `migrations/` are included.

## Verification

Run in the outputs directory: `go mod tidy`, `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean.
Smoke-tested `go run . --help` (start/migrate commands registered) and `start` with a missing
config file (fails fast with a wrapped error).
