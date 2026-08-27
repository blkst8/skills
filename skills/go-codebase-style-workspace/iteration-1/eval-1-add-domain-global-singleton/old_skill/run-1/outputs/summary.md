# Summary — `client` domain (Global Singleton pattern)

Added a complete `client` domain to the Go service following the blkst8 style
guide (`SKILL.md` + `references/style-guide.md`), using the **global
singleton** wiring (`app.A.*`) selected in the task.

## Requested deliverables

| Deliverable | File |
|---|---|
| Client model | `internal/models/client.go` |
| ClientRepository interface + sqlx impl (Create/Get/Update/Delete) | `internal/repository/client.go` |
| ClientUsecase with a Create flow | `internal/usecase/client.go` |
| Four Echo HTTP handlers (Create/Get/Update/Delete) | `internal/http/handlers/{create_client,get_client,update_client,delete_client}.go` |

## Supporting files (global singleton wiring + runnable slice)

- `internal/app/app.go` — `application` struct, `var A`, `init()`, `WithGracefulShutdown()`, `Wait()`
- `internal/app/db.go` — `WithDatabase()` mutates `A.Database` in place (global-pattern variant)
- `internal/app/repository.go` — `Repository` bundle (`Client repository.Client`), `WithRepository()`
- `internal/app/service.go` — `Service` bundle (`Client usecase.Client`), `WithService()`
- `internal/http/server.go` — `NewServer()` (no args) wires `POST/GET/PUT/DELETE /api/v1/clients` to the plain handler functions
- `internal/http/handlers/client_id.go` — shared `:id` path-param parser (returns 400 on bad input)
- `internal/log/log.go`, `internal/config/config.go`, `config.example.yaml` — zap + Viper per the guide
- `internal/http/middlewares/{logger,jwt_authentication}.go`, `internal/http/handlers/{healthz,metrics}.go` — referenced by the guide's canonical `server.go` template
- `cmd/root.go`, `cmd/start.go`, `main.go` — wiring order: `WithGracefulShutdown → WithDatabase → WithRepository → WithService → NewServer/Serve → Wait`
- `migrations/000001_create_clients.{up,down}.sql` — `clients` table matching the model
- `go.mod` / `go.sum` — module `yourproject` (placeholder; rename to your module path)

## Create flow (usecase)

`CreateClientInput{Name, Email}` → validate (required fields, `net/mail` address check; failures wrap the
`ErrInvalidClientInput` sentinel) → `repo.Create` (sqlx `NamedExecContext`, returns `LastInsertId`)
→ log `client created` (zap, structured) → `repo.Get` to return the stored record with generated ID and timestamps.

## Decisions

1. **ID type `uint32` everywhere** — the guide's model template uses `ID uint32` and its migration uses
   `INT AUTO_INCREMENT`, while its repo-interface sketch shows `id string`. I resolved the inconsistency in
   favor of `uint32` for type consistency; handlers parse `:id` with `strconv.ParseUint` and map bad input to 400.
2. **`repo.Create` returns `(uint32, error)`** — needed so the usecase can re-fetch the stored row and the
   handler can respond with the created record (the guide's handler template expects a result with an ID).
3. **Usecase lives in `internal/usecase`** (per the task's requested path) but is exposed through
   `app.A.Service.Client`, matching the guide's global-singleton access pattern exactly.
4. **Validation in the usecase** (business layer owns the rules); handlers only bind and map errors.
5. **Sentinel errors in the repository** — `ErrClientNotFound` (`sql.ErrNoRows` mapping, plus `RowsAffected == 0`
   for Update/Delete) and `ErrClientAlreadyExists` (MySQL error 1062 detection). Handlers map them to
   404/409 via `errors.Is` + `echo.NewHTTPError`, per the guide's HTTP error-handling pattern.
6. **Module path `yourproject`** is the guide's placeholder — rename in `go.mod` and imports when merging.

## Verification

- `go build ./...` — OK
- `go vet ./...` — OK
- `gofmt -l` — clean
