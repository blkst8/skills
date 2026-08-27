# Client domain — Global Singleton pattern (`app.A`)

Adds a `client` domain (model, sqlx repository, usecase, four Echo handlers)
to the existing Go service, wired through the `app.A` global singleton per the
blkst8 style guide. All handlers are plain functions accessing
`app.A.Usecase.Client.*`; no other pattern is mixed in.

## Files

| File | Purpose |
|---|---|
| `internal/models/client.go` | `Client` model — dumb struct with `db:` + `json:` tags, `*time.Time` for nullable `updated_at` |
| `internal/repository/client.go` | `Client` interface (Create/Get/Update/Delete) + private `client` sqlx impl + `NewClientRepository`; sentinel `ErrClientNotFound` (maps `sql.ErrNoRows`); named queries (`NamedExecContext`) for INSERT/UPDATE, `GetContext` for reads |
| `internal/usecase/client.go` | `ClientUsecase` interface + private `clientUsecase` + `NewClientUsecase`; owns the request DTOs (`CreateClientRequest`, `UpdateClientRequest`); Create flow persists and returns the row with the generated ID; Update writes then reads back; errors wrapped with `fmt.Errorf("failed to X: %w", err)` |
| `internal/http/handlers/create_client.go` | `POST` handler: bind → validate required fields → usecase → log → respond `CreateClientResponse` |
| `internal/http/handlers/get_client.go` | `GET :id` handler: parse path param → usecase → map `ErrClientNotFound` to 404 via `errors.Is` + `echo.NewHTTPError` |
| `internal/http/handlers/update_client.go` | `PUT :id` handler: parse id, bind, validate → usecase → 404 mapping → respond updated model |
| `internal/http/handlers/delete_client.go` | `DELETE :id` handler: parse id → usecase → 404 mapping → `204 No Content` |
| `internal/app/repository.go` | Global-singleton `Repository` bundle: `Client repository.Client` field; `WithRepository()` mutates `A` (uses `A.Database`) |
| `internal/app/usecase.go` | Global-singleton `Usecase` bundle: `Client usecase.ClientUsecase` field; `WithUsecase()` mutates `A` (uses `A.Repository.Client`) |
| `migrations/20260827000000_create_clients.{up,down}.sql` | `clients` table (id, name, email unique, created_at, updated_at) |

If the existing `internal/app/repository.go` / `usecase.go` already contain
other domains, merge the `Client` field and its construction line rather than
replacing the files.

## Wiring order (`cmd/start.go`, unchanged)

```go
app.WithGracefulShutdown()
app.WithDatabase()
app.WithRepository()  // now also constructs repository.NewClientRepository(A.Database)
app.WithService()
app.WithUsecase()     // now also constructs usecase.NewClientUsecase(A.Repository.Client)
```

## Routes to register in `internal/http/server.go`

```go
api := e.Group("/api/v1")
api.Use(middlewares.JWTAuthentication())
{
    api.POST("/clients", handlers.CreateClient)
    api.GET("/clients/:id", handlers.GetClient)
    api.PUT("/clients/:id", handlers.UpdateClient)
    api.DELETE("/clients/:id", handlers.DeleteClient)
}
```

## Style-guide conformance

- Layer chain: handlers → usecase → repository; no upward dependencies.
- `context.Context` first param on every I/O function.
- Interface + private impl + public constructor for repository and usecase.
- Imports grouped stdlib → external → internal with blank lines.
- snake_case files named after their action; receivers `c` (repo) and `u` (usecase).
- Only `log.Logger` (zap structured fields) — used in handlers; usecase logs nothing
  because its flows have no log-and-continue side effects.
- sqlx only: `NamedExecContext` for writes, `GetContext`/`ExecContext` for read/delete.

## Deliberate decisions

1. **`Create` takes `*models.Client` and populates `ID` via `res.LastInsertId()`**
   so the usecase's Create flow can return the real generated ID (the guide's
   snippet returns a zero ID otherwise).
2. **`Update`/`Delete` check `RowsAffected()` and return `ErrClientNotFound`**
   so `PUT`/`DELETE` on a missing row produce 404 instead of a silent 200.
3. **Request DTOs live in the usecase package** ("keep request/response DTO types
   next to the usecase that owns them"); handlers bind them directly. Response
   DTOs are handler-owned where a shape is constructed (`CreateClientResponse`);
   get/update return the model directly, delete returns 204.
4. **Module path placeholder `yourproject`** — replace with the real module path
   from `go.mod` (mechanical import rewrite).
5. No service layer client is involved: the client domain needs no external
   dependency service, so `ClientUsecase` depends only on the repository.
