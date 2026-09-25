# Summary

Built a Go service (`github.com/blkst8/client-service`) following the blkst8
Go codebase style guide (skill: `go-codebase-style`) that sends a Telegram
welcome message whenever a new client signs up — without ever failing client
creation when Telegram is down.

## Key deliverables

- `internal/service/notifier/notifier.go` — Telegram Bot API client:
  `Notifier` interface + private `telegramNotifier` implementation +
  `New(token)` constructor returning the interface. External I/O only
  (`POST /bot<token>/sendMessage` via `http.NewRequestWithContext`, 10s HTTP
  client timeout), errors wrapped with `fmt.Errorf("...: %w", err)`.
- `internal/usecase/client.go` — the client-creation flow: combines the
  `repository.Client` (data access) with the `notifier.Notifier` (external
  dependency). The welcome notification is a **best-effort side effect**:
  a notifier failure is logged with structured zap fields and the flow
  continues, so client creation still succeeds when Telegram is down.
  Repository errors remain fatal. Skips sending when no `telegram_id`.
- `internal/http/handlers/create_client.go` — binds/validates the request,
  calls the usecase, logs errors, maps sentinel errors to HTTP statuses.
  Plus `get_client.go` (404 via `repository.ErrClientNotFound` with
  `errors.Is`), `healthz.go`, and `handlers.go` (Handlers struct + `New(uc)`).

## Supporting files

- Wiring (**Private DI**, chosen as the guide's recommended default for
  production services and applied consistently): `internal/app/{app,db,
  repository,service,usecase}.go` bundles chained in `cmd/start.go`
  (signal-based graceful shutdown) with `cmd/root.go`, `main.go`.
- `internal/repository/client.go` — sqlx `NamedExecContext` insert (returns
  generated id) + `GetContext`, translating `sql.ErrNoRows` into the
  sentinel `ErrClientNotFound`.
- `internal/models/client.go` — dumb struct with `db:`+`json:` tags,
  `*time.Time` for nullable `updated_at`.
- `internal/config/config.go` + `config.example.yaml` — Viper config incl.
  new `telegram.token` section; `internal/log/log.go` — shared zap logger;
  `internal/http/server.go` + `middlewares/logger.go` — Echo server
  (`/api/v1/clients` routes, zap request logging, graceful shutdown).
- `migrations/20260827000000_create_clients.{up,down}.sql` — clients table
  with `telegram_id` column.
- Tests: `internal/usecase/client_test.go` (table-driven, fakes for repo +
  notifier — proves "Telegram down ⇒ creation still succeeds"), and
  `internal/service/notifier/notifier_test.go` (httptest server asserting
  payload/headers and provider-error mapping).
- `Makefile`, `README.md`, `go.mod`/`go.sum`.

## Decisions

- **Private DI wiring** (instead of global singleton) — production service,
  testability is required for the fake-based usecase test; pattern applied
  consistently everywhere, never mixed.
- Usecase owns `CreateClientRequest`; the handler binds its own wire DTO and
  maps explicitly into the usecase type.
- `repository.Client.Create` returns `(uint32, error)` so the generated id
  flows back into the created model (small extension of the reference
  pattern; reference returns only `error`).
- Telegram send is skipped entirely when `telegram_id` is empty; message
  text is composed in the usecase (`Welcome, <name>!`).
- No JWT/auth middleware or metrics — out of scope for this task; routes
  registered match implemented handlers so everything compiles and runs.

## Verification

`go build ./...`, `go vet ./...`, `gofmt -l` (clean), and `go test ./...`
all pass — including the `succeeds_even_when_telegram_is_down` subtest.
