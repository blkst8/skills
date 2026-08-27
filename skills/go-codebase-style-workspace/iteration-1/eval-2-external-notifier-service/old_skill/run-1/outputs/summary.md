# Summary — Telegram welcome-message notifier for the client-creation flow

Task: add a Telegram Bot API notifier client and wire a best-effort welcome
message into client creation, so that client creation still succeeds when
Telegram is down. Everything follows the blkst8 Go codebase style guide.

## What was built

Module `github.com/blkst8/client-service` — a complete, compiling, tested Go
service (Echo + zap + Viper + Cobra + sqlx/MySQL):

```
main.go
config.example.yaml, Makefile, README.md, go.mod, go.sum
migrations/000001_init.{up,down}.sql          # clients table incl. telegram_id
cmd/root.go, cmd/start.go                     # Cobra commands; start.go wires DI
internal/config/config.go                     # Viper config; new Telegram section
internal/log/log.go                           # zap setup (log.Logger)
internal/models/client.go                     # Client model (db+json tags)
internal/repository/client.go                 # interface + MySQL impl, ErrClientNotFound
internal/services/notifier/notifier.go        # Notifier interface + New() + noop
internal/services/notifier/telegram.go        # Telegram Bot API client
internal/services/client.go                   # ClientService: Create + best-effort welcome
internal/app/{app,db,repository,service}.go   # private-DI bundles (no globals)
internal/http/server.go                       # Echo server (private-DI variant)
internal/http/handlers/{handlers,create_client,get_client,healthz}.go
internal/http/middlewares/logger.go           # ZapLogger middleware
internal/services/{client_test.go}            # table-driven tests incl. "Telegram down"
internal/services/notifier/telegram_test.go   # httptest-based Telegram API tests
```

## Key decisions

- **Wiring pattern (skill Step 0): Private DI.** The session is non-interactive
  so the user could not be asked; the guide recommends Private DI for
  production/testable services, and the resilience requirement is best
  demonstrated with injected interfaces. Applied consistently: no global
  `app.A`, `cmd/start.go` threads `WithDatabase → WithRepository → WithServices`
  into `httpserver.NewServer(svc)`.
- **Notifier placement.** Per the guide, the `Service` bundle
  (`internal/app/service.go`) holds service instances *and shared infra
  clients*, so the notifier lives there (`Service.Notifier`) and is injected
  into `ClientService`. The client itself is interface-driven like the
  repositories: `Notifier` interface + private `telegram` struct + `New()`
  constructor in `internal/services/notifier/`.
- **Resilience ("Telegram down → creation still succeeds").** `ClientService.Create`
  commits the client first, then calls `sendWelcomeMessage`, which logs the
  notifier error with zap and never propagates it. Defense in depth:
  `http.Client` per-call timeout (default 5s, configurable) bounds the request,
  and `telegram.enabled: false` wires a no-op notifier.
- **Telegram client details.** POST `{api_base}/bot<token>/sendMessage` with
  JSON `{chat_id, text}`; non-2xx, transport failure, or `ok: false` becomes a
  wrapped `fmt.Errorf(... %w)` error; response body size capped.
- **Deliberate deviation:** repository `Create` takes `*models.Client` (not a
  value) so it can write the generated `LastInsertId` back — the guide's own
  handler example returns an ID from the service, which a value-based
  `Create ... error` cannot provide.
- Scope notes: JWT middleware omitted (not requested, auth unspecified);
  worker framework not included (no background jobs in this flow); the
  notifier was placed in `internal/services/notifier` per the guide's
  `internal/services` convention rather than the illustrative
  `internal/service|usecase` paths in the task examples.

## Verification

- `go build ./...`, `go vet ./...`, `gofmt` — clean
- `go test ./...` — all tests pass, including the two critical cases:
  - `TestCreate/"telegram down does not fail creation"` — notifier error is
    logged, `Create` still returns the stored client
  - `TestSendMessage` — validates request path/payload and error mapping for
    API rejections (simulated outage via `httptest`)
