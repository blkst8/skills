# JWT auth + zap request-logging middleware for Echo

Built per the `go-codebase-style` skill (SKILL.md + `http-workers.md`,
`config-logging-db.md`, `structure-conventions.md`, `wiring-patterns.md`).

## Files

| File | Purpose |
|---|---|
| `internal/http/middlewares/jwt_authentication.go` | `JWTAuthentication(secret)` — Bearer-token validation with `golang-jwt/jwt/v5`, stores claims in context under `"user"` |
| `internal/http/middlewares/logger.go` | `ZapLogger(logger, skipURLs...)` — one structured zap entry per request, skips the given paths |
| `internal/http/server.go` | Echo server: middleware chain, public + protected routes, `Serve`/`Shutdown` |
| `internal/config/config.go` | Viper config structs (skill reference + a new `JWT` section for the signing secret) |

## Registration in `internal/http/server.go`

```go
e.Use(middleware.Recover())
e.Use(middleware.RequestID())
e.Use(middlewares.ZapLogger(log.Logger, "/healthz", "/metrics")) // skip list here
e.Use(middleware.CORS())

e.GET("/healthz", handlers.Healthz)   // public, not logged
e.GET("/metrics", handlers.Metrics)   // public, not logged

api := e.Group("/api/v1")
api.Use(middlewares.JWTAuthentication(config.C.JWT.Secret)) // protected group
```

## Decisions

1. **Wiring pattern: Private DI.** The skill says to ask once (Step 0); since
   this is a non-interactive run I used the skill's stated default for
   production-bound services: `NewServer(uc *app.Usecase)` with
   `handlers.New(uc)`. The Global Singleton variant differs only in the
   `NewServer()` signature and using package-level handlers — the middleware
   and registration lines are identical.
2. **`JWTAuthentication` takes the secret as a parameter.** The skill's JWT
   snippet is a skeleton (`// Validate token`) and can't validate anything
   without a key. Following the codebase's own pattern of middlewares
   receiving their dependencies as arguments (`ZapLogger(log.Logger, ...)`),
   the secret is injected from `config.C.JWT.Secret` in `server.go` rather
   than read from a global inside the middleware, keeping it testable.
3. **Logger skip list is path-based via a map** built once from the
   variadic `skipURLs`; skipped requests bypass timing/logging entirely.
   Added an `echo.HTTPError`-aware status correction so failed requests log
   their real status (Echo's error handler runs after middleware unwinds).
4. **JWT algorithm allow-list** (`HS256/384/512` via `jwt.WithValidMethods`)
   instead of trusting the token header; auth failures return
   `echo.ErrUnauthorized` without leaking validation details.
5. **`package http`** kept as in the skill's `server.go` examples; when
   imported elsewhere (e.g. `cmd/start.go`) alias it as `httpserver` per the
   "avoid generic names" convention.
6. Module path is the placeholder `yourproject` — replace with the real
   module path. Sibling packages referenced but out of scope here:
   `internal/http/handlers` (`New(uc)`, `Healthz`, `Metrics`, client CRUD)
   and `internal/app` (`Usecase` bundle).
