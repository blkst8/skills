# Layout Reference

What belongs in every directory of the monorepo, and the conventions inside
each app.

## Top level

```
my-project/
├── apps/               # deployable applications
├── packages/           # shared libraries
├── infra/              # docker/, nginx/, terraform/
├── scripts/            # dev.sh, build.sh, seed-db.sh
├── docs/               # architecture.md, decisions
├── .github/workflows/  # CI
├── docker-compose.yml
├── pnpm-workspace.yaml
├── turbo.json
├── tsconfig.base.json
├── .env.example
├── .gitignore
├── package.json
└── README.md
```

## apps/

Deployable applications only. Each must be independently runnable,
buildable, testable, and deployable — that independence is the whole point
of the monorepo, so protect it: no app imports another app's source.

- `apps/web` — React, Next.js, Vue, Nuxt, Angular, ...
- `apps/api` — Express, NestJS, Fastify, Go, Python, ...

Naming by technology (`web`/`api`) beats naming by framework (`nextjs`/
`nestjs`) — frameworks get swapped more often than roles do.

More apps are fine: `apps/admin`, `apps/worker`, `apps/mobile`.

## packages/

Shared libraries. Each gets its own `package.json` and a scoped name:

```json
{
  "name": "@repo/types",
  "version": "0.0.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts"
}
```

Pointing `main`/`types` directly at source (`./src/index.ts`) lets
consumers compile it themselves — no build step needed for internal
packages, which removes a whole class of staleness bugs.

### packages/types — shared types and DTOs

```
packages/types/src/
├── user.ts
├── product.ts
├── api.ts        # request/response envelopes
└── index.ts      # re-export everything
```

Types only — `import type` territory. No runtime code. If it executes,
it belongs in `packages/shared`.

```ts
export interface UserDto {
  id: string;
  email: string;
  name: string;
}
```

### packages/shared — shared runtime logic

```
packages/shared/src/
├── date.ts          # date helpers
├── validation.ts    # zod / class-validator schemas
└── constants.ts
```

Validation schemas are the highest-value resident: the API validates a
request with the same schema the frontend validates its form with, so the
two can never drift apart.

### packages/ui — shared components (only when earned)

```
packages/ui/src/
├── Button.tsx
├── Modal.tsx
└── Input.tsx
```

Create it when a second frontend app appears or the team commits to a
design system. With one frontend app, components live in
`apps/web/src/components/`.

### packages/config — shared tooling config

```
packages/config/
├── eslint/         # shared ESLint config
├── typescript/     # shared tsconfig presets
├── prettier/
└── package.json
```

Exposes configs like `@repo/config/typescript/base.json`. Apps extend
instead of copy, so a lint rule or compiler option changes once.

## Inside apps/api — feature modules

```
apps/api/src/
├── modules/
│   ├── auth/
│   │   ├── auth.controller.ts
│   │   ├── auth.service.ts
│   │   ├── auth.routes.ts
│   │   └── auth.types.ts
│   ├── users/
│   │   ├── users.controller.ts
│   │   ├── users.service.ts
│   │   ├── users.repository.ts
│   │   └── users.types.ts
│   └── products/
├── common/
│   ├── middleware/
│   ├── errors/
│   ├── guards/
│   └── utils/
├── config/
├── database/        # schema, migrations, connection
└── main.ts
```

One module per business capability. A change to "auth" touches one folder.
Cross-cutting code (middleware, error handling) goes in `common/`, not into
a feature.

Adapt the file names to the framework — NestJS uses `*.module.ts` +
`*.dto.ts`, Fastify uses plugins, Go uses a package per module — but keep
the feature-grouping idea.

## Inside apps/web — feature folders

```
apps/web/src/
├── app/               # routes/pages (Next.js) or app shell
├── features/
│   ├── auth/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── api.ts
│   │   └── types.ts
│   ├── users/
│   └── products/
├── components/
│   ├── layout/
│   └── common/        # generic app-specific components
├── lib/               # app-level helpers
├── config/
└── styles/
```

Four tiers of components, so ownership never gets murky:

| Tier | Lives in | Scope |
|---|---|---|
| Design-system components | `packages/ui` | Reusable across apps |
| Generic app components | `apps/web/src/components/common` | One app |
| Business components | `apps/web/src/features/<f>/components` | One feature |
| Layout | `apps/web/src/components/layout` | App chrome |

## Environment variables

```
my-project/
├── .env.example          # documents the union, points to per-app files
├── apps/web/
│   ├── .env.local
│   └── .env.example
└── api/
    ├── .env
    └── .env.example
```

`apps/web/.env.local`:
```
NEXT_PUBLIC_API_URL=http://localhost:4000
```

`apps/api/.env`:
```
DATABASE_URL=postgresql://user:password@localhost:5432/mydb
JWT_SECRET=secret
PORT=4000
```

Never commit real `.env` files — `.gitignore` them, commit `.env.example`
with placeholder values so newcomers know what to fill in.

## Docker

- `docker-compose.yml` at repo root: local infrastructure (Postgres, Redis).
- `apps/<app>/Dockerfile`: one per containerized app, so each deploys alone.

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: app
      POSTGRES_DB: app
    ports:
      - "5432:5432"
  redis:
    image: redis:7
    ports:
      - "6379:6379"
```

`infra/` holds deployment-adjacent config that isn't needed for local dev:
nginx configs, terraform, prod Dockerfiles. Keep it out of the apps so
infrastructure changes never force app rebuilds.
