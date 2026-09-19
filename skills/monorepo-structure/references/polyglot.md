# Polyglot Monorepos (Go, Python, and other non-JS backends)

Same layout, different rules. The `apps/` + `packages/` shape survives
unchanged; what changes is how tooling and sharing work, because the JS
toolchain can only see JS.

## Core rule: each language stays native

- `apps/web` — normal JS citizen: `package.json`, pnpm workspaces, Turbo
  tasks.
- `apps/api` — a native module of its language: `go.mod` for Go,
  `pyproject.toml` for Python. No `package.json`, no TypeScript, no pnpm
  inside it.
- `packages/` — JS packages (`types`, `shared`, `config`) remain JS-only
  and serve JS consumers. Don't create a "shared" Go/Python package here
  unless two non-JS apps need it; Go prefers intra-module packages, Python
  prefers a `libs/` dir.

Forcing pnpm onto Go code (wrapping every `go test` in npm scripts inside
the Go app) makes both ecosystems harder to use. Native tools stay native;
orchestration happens one level up.

```
my-project/
├── apps/
│   ├── web/               # React/Vue — package.json, pnpm
│   └── api/               # Go — go.mod, cmd/, internal/
│       ├── cmd/api/main.go
│       ├── internal/
│       │   ├── auth/
│       │   ├── users/
│       │   └── store/     # db access lives here (Prisma equivalent)
│       └── go.mod
├── packages/
│   ├── types/             # TS types for JS consumers
│   └── shared/
├── pnpm-workspace.yaml    # only lists JS workspaces — that's fine
├── docker-compose.yml
└── Makefile               # cross-language tasks
```

Note `apps/api/internal/` — Go's feature-grouping equivalent, and it
carries the same meaning as `modules/`: `auth/`, `users/`, `store/`.
Database access stays in the backend exactly as in the TS world.

## Cross-language sharing: schemas, not types

TypeScript types can't be imported by Go, so `packages/types` doesn't span
the language gap. Share the *contract* instead:

- **OpenAPI spec** — generate TS types for the frontend from the API spec
  (`openapi-typescript`), or generate both server stubs and client types
  from one source. Best default: the API is the source of truth.
- **Protobuf / JSON Schema** — when RPC or schema-first workflows are
  already in play.

Put the contract source (or the generated artifacts for JS) where both
sides can reach it: `packages/api-schema/` or `docs/api/`. The principle
from SKILL.md carries over — frontend knows DTOs, never database internals
— but the DTOs travel as generated code, not hand-written imports.

Validation is shared the same way: a JSON Schema or OpenAPI validation
keywords, with each side generating its own runtime validator. Don't
hand-port a zod schema to Go; generate both from one spec.

## Task orchestration

Turbo only runs tasks in packages that have a `package.json`, so the Go app
is invisible to it. Two good options:

**Makefile at the root** (simplest, language-agnostic):

```makefile
dev:        ## web + api
	pnpm --filter web dev &; go run ./apps/api/cmd/api

build:
	pnpm turbo build; cd apps/api && go build -o bin/api ./cmd/api

test:
	pnpm turbo test; cd apps/api && go test ./...
```

**A thin `package.json` wrapper** (keeps `pnpm dev`/`pnpm test` working) —
apps/api gets a minimal `package.json` whose scripts call native tools:

```json
{
  "name": "api",
  "private": true,
  "scripts": {
    "build": "go build ./cmd/api",
    "test": "go test ./..."
  }
}
```

Turbo then schedules it like any other workspace. This is fine *because* the
JS in this file is pure delegation — zero JS dependencies in the Go app.

Pick one and tell the user which commands are canonical; two orchestration
layers confuse newcomers.

## Environment, Docker, CI

- Per-app envs still apply: `apps/web/.env.local`, `apps/api/.env` (or
  `.env.example` — Go/Python read them with their own libraries).
- Docker-compose at root for infrastructure; `apps/api/Dockerfile` is a
  multi-stage Go build, independent of the web image.
- CI gets a job per language: `web` job runs pnpm/turbo steps, `api` job
  runs native steps. Keep them separate — one setup-python shouldn't
  download node_modules.

## What changes vs the TS stack, in short

| Concern | TS-only | Polyglot |
|---|---|---|
| Workspace | pnpm everywhere | pnpm for JS side only |
| Shared types | `@repo/types` imports | Generated from OpenAPI/proto |
| Shared validation | zod in `@repo/shared` | Generated from schema |
| Tasks | turbo | Makefile or thin wrappers |
| DB code | `apps/api/src/database/` | `apps/api/internal/store/` (same rule) |
