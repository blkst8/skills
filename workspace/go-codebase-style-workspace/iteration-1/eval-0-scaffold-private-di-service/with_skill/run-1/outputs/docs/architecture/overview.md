# Architecture Overview

Layer chain (dependencies never point upward):

```
HTTP handlers (internal/http/handlers)        Job/task handlers (internal/worker/handlers, internal/workers)
        └───────────────┬──────────────────────────────────┘
                        ▼
              Usecase (internal/usecase)          ← big business flows
              ┌──────┴───────┐
              ▼              ▼
      Repository         Service (<role>)
   (internal/repository) (internal/service/notifier) ← external I/O only
              ▼
           Models (internal/models)
              ▼
        Database (internal/app/db.go)
```

## Wiring: private dependency injection

This service uses the **private DI** pattern (`references/wiring-patterns.md`,
pattern 2). `internal/app/` holds typed bundles and constructors, no globals:

- `app.WithDatabase() *sqlx.DB`
- `app.WithRepository(db) *Repository`
- `app.WithServices() *Service` (external clients only)
- `app.WithUsecases(repo, svc) *Usecase`

`cmd/start.go` chains them and passes `*app.Usecase` into the HTTP server,
the ticker workers and the worker pool. Config (`config.C`) and logging
(`log.Logger`) remain package-level, matching the style guide.

## Background jobs

- `internal/worker` (ticker): periodic reconcile job, interval from
  `worker.jobs_intervals.reconcile_invoices` (5m).
- `internal/workers` (pool): concurrent on-demand tasks; enabled via
  `workers.enabled`.
