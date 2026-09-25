# invoice-service scaffold — summary

Scaffolded a complete Go service (`github.com/blkst8/invoice-service`) per the
**go-codebase-style** skill (v1.1), using the **private dependency injection**
pattern (explicitly requested). `go build ./...`, `go vet ./...` and `gofmt`
all pass; `go.mod`/`go.sum` were generated with `go mod tidy`.

## What was built

- **Wiring (`internal/app/` + `cmd/start.go`)** — private DI bundles:
  `WithDatabase() *sqlx.DB` → `WithRepository(db) *Repository` →
  `WithServices() *Service` (external clients only) →
  `WithUsecases(repo, svc) *Usecase`. `cmd/start.go` chains them, starts the
  ticker worker, the optional worker pool, the Echo server, and blocks on a
  `signal.NotifyContext` for graceful shutdown (`defer db.Close()`,
  `reconcileJob.Close()`, `pool.Stop()`, `srv.Shutdown()`).
- **MySQL** — `internal/app/db.go` (sqlx + pooling + Ping), initial migration
  `migrations/20260827000000_create_invoices.{up,down}.sql`,
  `cmd/migrate.go` (golang-migrate up/down), `make migrate-up/down`.
- **REST API (Echo)** — `internal/http/server.go` with
  Recover/RequestID/ZapLogger/CORS middleware, `/healthz`, `/metrics`
  (Prometheus via `internal/metrics`), and a JWT-protected `/api/v1` group:
  `POST /invoices`, `GET /invoices`, `GET /invoices/:id`.
- **Reconcile background job** — ticker worker (`internal/worker`) running
  `worker/handlers/reconcile_invoices.go` every
  `worker.jobs_intervals.reconcile_invoices` (default `5m`): lists pending
  invoices, marks past-due ones `overdue`, emits Prometheus counters and
  notifies ops through the webhook notifier. Also exposed as an on-demand
  task for the worker pool (`internal/workers`, disabled by default).
- **Layers** — interface + private impl + public constructor everywhere:
  `repository.Invoice` (sqlx, named inserts, sentinel errors
  `ErrInvoiceNotFound`/`ErrInvoiceNumberTaken` from `sql.ErrNoRows` and MySQL
  error 1062), `service/notifier` (external webhook I/O only, no DB),
  `usecase.InvoiceUsecase` (combines repo + notifier, owns request DTOs,
  log-and-continue for non-fatal side effects), `handlers.Handlers` (bind →
  usecase → log → `errors.Is` → `echo.NewHTTPError` mapping), dumb
  `models.Invoice` with `db:`/`json:` tags and pointer nullable field.
- **Support files** — viper config (`config.go` + `builtin.go` defaults,
  `config.example.yaml`), zap logging (`internal/log`), Makefile with
  ldflags build-info injection (`internal/app` vars + `Banner()`), systemd
  unit, README, `docs/` notes, `.gitignore` (excludes real `config.yaml`).

## Decisions / deviations worth noting

1. **Config struct tags**: kept the guide's `yaml:` tags and added
   `mapstructure:` tags — viper.Unmarshal uses mapstructure, so multi-word
   keys like `read_timeout` would otherwise silently decode to zero values.
2. **`repository.Invoice.Create` returns `(uint32, error)`** (last insert ID)
   instead of the guide's skeleton `error`-only form, so the usecase can
   return the created invoice with a real ID.
3. **Notifier no-op**: with an empty `notifier.webhook_url` the webhook
   notifier is disabled (`Send` returns nil), so the scaffold runs without
   external infrastructure.
4. **JWT middleware** included per the guide's server setup and essential
   dependencies list (golang-jwt/v5, HMAC secret from `jwt.secret`).
5. **`log.SetLevel`** helper added so `logger.level` from config actually
   applies (the guide's log.go hard-codes the production config).
6. All naming/import/error conventions applied: snake_case action-named
   files, 1–2 letter receivers, stdlib→external→internal imports,
   `context.Context` first param on every I/O function, `fmt.Errorf("...: %w", err)`
   wrapping, zap-only logging.
