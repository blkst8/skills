# JWT Authentication + Zap Request-Logging Middleware (Echo)

Implements the two middlewares and their registration in `internal/http/server.go`
following the `go-codebase-style` skill / style guide.

## Files

| File | Purpose |
|---|---|
| `internal/http/middlewares/jwt_authentication.go` | JWT auth middleware, `Claims` type, `ClaimsFromContext` helper, `validateToken` |
| `internal/http/middlewares/logger.go` | Zap request-logging middleware with variadic skip-URL list |
| `internal/http/server.go` | Echo server: global middleware registration, public routes, JWT-protected `/api/v1` group |

## Conventions applied (from the style guide)

- Files are snake_case (`jwt_authentication.go`, `logger.go`); package `middlewares`
  (lowercase, plural for collections) under `internal/http/middlewares/`.
- Both middlewares are constructors returning `echo.MiddlewareFunc`, matching the
  guide's signatures exactly: `ZapLogger(log *zap.Logger, skipURLs ...string)` and
  `JWTAuthentication()`.
- Imports grouped stdlib → external → internal (blank line between groups).
- Package- and function-level godoc per the Documentation section.
- Logging fields are structured zap fields only; no `fmt.Println` / stdlib `log`.
- Module path `yourproject` is the guide's placeholder.

## Registration in `server.go`

- `e.Use(middlewares.ZapLogger(log.Logger, "/healthz", "/metrics"))` — registered
  globally with `e.Use()` **before** any route is defined, so it wraps everything,
  while the skip list keeps health checks and metrics scraping out of the logs.
- `api.Use(middlewares.JWTAuthentication())` — registered on the `/api/v1` group,
  exactly as the style guide's `server.go` template shows, so `/healthz` and
  `/metrics` stay unauthenticated.
- Global middleware stack follows the guide's template order: `Recover`,
  `RequestID`, `ZapLogger`, `CORS`.

## Decisions

1. **Private DI wiring** for `server.go` (`NewServer(svc *app.Service)`, one
   `*handlers.Handlers` built via `handlers.New(svc)`). The guide recommends this
   pattern for production services. Under the global-singleton pattern the
   middleware code and registration lines are identical; only the handler wiring
   changes (plain functions using `app.A`, `NewServer()` without args).
2. **`echo.ErrUnauthorized` for all JWT failures** (missing header, non-Bearer,
   bad signature/expiry, wrong signing method) per the guide's middleware template.
3. **JWT secret from config**: `JWTAuthentication()` reads
   `config.C.HTTPServer.JWTSecret`. This requires one addition to the
   `HTTPServer` config struct: `JWTSecret string \`yaml:"jwt_secret"\`` (and a
   matching entry in `config.example.yaml` — never commit the real secret).
   Only HMAC signing methods are accepted; `alg=none` is rejected.
4. **Status-derivation fix in `ZapLogger`**: Echo writes error responses only
   after the middleware chain unwinds, so `ctx.Response().Status` still reads
   `200` when a handler returns e.g. `echo.ErrUnauthorized`. The middleware
   derives the status from the returned `*echo.HTTPError` (falling back to 500),
   mirroring Echo's built-in logger. Verified by smoke test (401s log as 401).
5. **`Claims` + `ClaimsFromContext`**: claims are stored in the Echo context
   under the `"user"` key (per the guide); the helper lets protected handlers
   retrieve them.

## Verification

- `go build ./...`, `go vet ./...`, `gofmt` clean against a scratch module with
  echo v4.15.4, golang-jwt/jwt v5.3.1, zap v1.28.0.
- Runtime smoke test against a live server:
  - `GET /healthz`, `GET /metrics` → 200, **not** logged (skip list works).
  - `GET /api/v1/...` without token / garbage token / `alg=none` token → 401.
  - Valid HS256 token → 200; logged line shows `method`, `uri`, `status`,
    `latency` with status reported accurately.
