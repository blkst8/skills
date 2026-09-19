---
name: monorepo-structure
description: >
  Scaffold, restructure, audit, or work within a frontend + backend monorepo
  using the apps/ + packages/ layout. Default stack is pnpm workspaces +
  Turborepo for JS/TS repos, with adaptations for polyglot repos where the
  backend is Go, Python, or another non-JS language. Use this skill whenever
  the user mentions a monorepo, workspaces, pnpm, turborepo, wants frontend
  and backend in the same repo, asks to restructure or modernize a project
  into apps/web + apps/api, wants to share types, DTOs, validation schemas,
  or utilities between client and server, or asks where a piece of shared
  code belongs. Also use when reviewing an existing repo's structure, when a
  repo has tight coupling between frontend and backend, or when starting any
  new full-stack project from scratch — even if the word "monorepo" is never
  said out loud.
---

# Monorepo Structure

A monorepo exists to make five things easy: sharing code between apps,
running/building/testing each app independently, keeping environment configs
clean, scaling as the project grows, and avoiding tight coupling between
frontend and backend. Every recommendation in this skill serves at least one
of those five goals. When you improvise, check your choice against them.

## What kind of request is this?

Classify the request before doing anything:

| The request looks like... | What to do |
|---|---|
| "Set up a new project", greenfield, blank repo | Scaffold — this file plus the two reference files for the chosen stack |
| "Restructure / clean up / convert our repo into a monorepo" | Read `references/migration.md` first — restructuring has failure modes scaffolding doesn't |
| "Where should X go?", "should this be shared?" | Answer from the principles below; no scaffolding needed |
| "Review our structure", "is our repo set up well?" | Audit checklist in `references/migration.md` |

Mixed requests are common ("set up a monorepo and move our existing code
in") — handle each part with its section.

## The layout at a glance

```
my-project/
├── apps/
│   ├── web/          # frontend — Next.js, React, Vue, ...
│   └── api/          # backend — NestJS, Express, Fastify, Go, Python, ...
├── packages/
│   ├── types/        # shared types / DTOs
│   ├── shared/       # shared utilities, validation, constants
│   └── config/       # shared TS/ESLint/Prettier config
├── infra/            # docker, nginx, terraform
├── scripts/          # dev.sh, build.sh, seed-db.sh
├── docs/
├── docker-compose.yml
├── pnpm-workspace.yaml
├── turbo.json
├── tsconfig.base.json
└── package.json
```

Two rules define the layout:

- **`apps/` holds deployable applications.** Everything in `apps/` can be
  built, run, and deployed on its own. If it can't ship independently, it
  isn't an app.
- **`packages/` holds shared libraries.** If two apps both need it and it's
  stable, it's a package. If only one app uses it, keep it inside that app —
  premature sharing is coupling.

Add `packages/ui` (shared React components) only when there is more than one
frontend app or a deliberate design system. One frontend app does not need it.

For what belongs in each directory, per-package `package.json` conventions,
and feature-based structure inside `apps/web` and `apps/api`, read
`references/layout.md`.

## Choose the tooling path

Ask (or detect from the repo) one question: **is every app JavaScript or
TypeScript?**

- **All JS/TS** → pnpm workspaces + Turborepo. Read `references/ts-stack.md`
  for the root `package.json`, `pnpm-workspace.yaml`, `turbo.json`, and
  `tsconfig.base.json` that make the layout work.
- **Polyglot** (Go, Python, or other backend) → keep the same `apps/` +
  `packages/` shape, but the JS workspace governs only the JS side; the
  backend stays a native module. Read `references/polyglot.md` — forcing pnpm
  or TypeScript conventions onto non-JS code makes both sides harder to work
  with.
- **Nx** is a legitimate alternative when the user wants generators and a
  dependency graph and doesn't mind opinionated tooling. Don't push it; the
  pnpm + Turborepo default is simpler, faster to set up, and easier to
  reason about.

If the user already committed to a different tool (yarn workspaces, Nx,
bazel), work within it — the layout and principles below still apply.

## Principles that outlive any template

These apply to every scenario: scaffolding, migrating, answering "where
does this go?". Understanding *why* matters more than memorizing the tree.

### 1. Share only stable code

Good candidates for `packages/`: types, DTOs, validation schemas,
constants, utility functions, config, UI components.

Don't share: backend services, database models, app-specific business
logic. Sharing a database model drags the ORM, connection handling, and
backend deployment concerns into every consumer — the moment the backend
schema changes, everything that imports it breaks. Expose a DTO through
`packages/types` instead.

Rule of thumb: packages should change *slower* than the apps that depend on
them. If a "shared" thing changes every sprint, it belongs in the app that
owns it.

### 2. Frontend never imports backend source

Good:
```ts
import type { UserDto } from "@repo/types";
```

Bad:
```ts
import { UserService } from "../../api/src/modules/users/users.service";
```

Deep relative imports across app boundaries work until someone moves a
file, and they hide which app actually depends on which. Workspace package
imports (`@repo/types`) make the dependency explicit and survive refactors.

### 3. Database code lives in the backend

`apps/api/src/database/` holds schema, models, and migrations. The frontend
knows the world through DTOs in `packages/types`, never through database
internals. This keeps the frontend deployable without the backend's
toolchain and lets the database evolve without frontend rebuilds.

### 4. Group by feature, not by technical layer

Inside both `apps/api` and `apps/web`, keep a feature's files together:

```
apps/api/src/modules/
├── auth/          # auth.controller.ts, auth.service.ts, auth.types.ts
└── users/         # users.controller.ts, users.service.ts, users.repository.ts
```

Layer-based grouping (`controllers/`, `services/`, `repositories/`) puts a
single change three directories away from itself. Feature grouping keeps
related files close so changes stay local — the same reason apps/ and
packages/ exist at the top level.

### 5. One `.env` per app, never committed

```
apps/web/.env.local      # NEXT_PUBLIC_API_URL=http://localhost:4000
apps/api/.env            # DATABASE_URL=..., JWT_SECRET=..., PORT=4000
```

A single repo-wide `.env` means every environment holds secrets it doesn't
need, and one deployable can't be configured independently. Commit only
`.env.example` files; real env files stay out of git.

### 6. Docker-compose at the root, Dockerfiles in the apps

`docker-compose.yml` at the repo root for local infrastructure (Postgres,
Redis). When containerizing the apps themselves, each app gets its own
`apps/<app>/Dockerfile` — independently deployable means independently
containerizable.

## Answering "where does X go?"

Placement questions deserve short, justified answers — resist the urge to
scaffold anything. Run the item through the share/avoid lists in Principle 1
and name the destination:

- "shared validation schema" → `packages/shared` (runtime code, not just
  types — doesn't belong in `packages/types`)
- "DTO the frontend consumes" → `packages/types`
- "Prisma/Drizzle models" → stays in `apps/api`; expose types via DTOs
- "helper only the web app uses" → `apps/web/src/lib/`, not `packages/shared`

When the user's instinct contradicts the principles (they want to share
database models, say), explain the cost using the reasoning above rather
than just refusing — the coupling argument usually lands.

## Reference files

Read these on demand — don't load them all up front:

- `references/layout.md` — what goes in every directory, package.json
  conventions, feature-based structure for web and api
- `references/ts-stack.md` — pnpm + Turborepo configuration, file by file
- `references/polyglot.md` — Go/Python/other backends, cross-language
  sharing, task orchestration
- `references/migration.md` — audit checklist and step-by-step restructure
  of an existing repo
