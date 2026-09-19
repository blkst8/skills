A good **monorepo structure** for a project with both **frontend** and **backend** should make it easy to:

- share code between apps
- run/build/test each app independently
- keep environment configs clean
- scale as the project grows
- avoid tight coupling between frontend and backend

A common and practical structure is:

````markdown
my-project/
├── apps/
│   ├── web/
│   │   ├── src/
│   │   ├── public/
│   │   ├── package.json
│   │   └── ...
│   │
│   └── api/
│       ├── src/
│       ├── tests/
│       ├── package.json
│       └── ...
│
├── packages/
│   ├── shared/
│   │   ├── src/
│   │   └── package.json
│   │
│   ├── ui/
│   │   ├── src/
│   │   └── package.json
│   │
│   ├── config/
│   │   ├── eslint/
│   │   ├── typescript/
│   │   └── package.json
│   │
│   └── types/
│       ├── src/
│       └── package.json
│
├── infra/
│   ├── docker/
│   ├── nginx/
│   └── terraform/
│
├── scripts/
│   ├── dev.sh
│   ├── build.sh
│   └── seed-db.sh
│
├── docs/
│   └── architecture.md
│
├── .github/
│   └── workflows/
│
├── .env.example
├── package.json
├── pnpm-workspace.yaml
├── turbo.json
├── tsconfig.base.json
├── eslint.config.js
└── README.md
````

## Recommended layout

### `apps/`

Put deployable applications here.

```text
apps/
├── web/
└── api/
```

Examples:

- `apps/web` — React, Next.js, Vue, Nuxt, Angular, etc.
- `apps/api` — Express, NestJS, Fastify, Django, Laravel, Spring Boot, etc.

Each app should be independently runnable and deployable.

---

## `packages/`

Put shared libraries here.

```text
packages/
├── shared/
├── ui/
├── types/
└── config/
```

### Common packages

#### `packages/shared`

Shared utility logic.

```text
packages/shared/
├── src/
│   ├── date.ts
│   ├── validation.ts
│   └── constants.ts
└── package.json
```

Use this for things like:

- constants
- validation schemas
- date helpers
- shared business utilities
- API response helpers

---

#### `packages/types`

Shared TypeScript types.

```text
packages/types/
├── src/
│   ├── user.ts
│   ├── product.ts
│   └── api.ts
└── package.json
```

Good for:

- DTOs
- API request types
- API response types
- shared domain models

Example:

```ts
export interface UserDto {
  id: string;
  email: string;
  name: string;
}
```

---

#### `packages/ui`

Shared frontend UI components.

```text
packages/ui/
├── src/
│   ├── Button.tsx
│   ├── Modal.tsx
│   └── Input.tsx
└── package.json
```

Use this if you have more than one frontend app, or if you want a clean design system.

If you only have one frontend app, you can skip this initially.

---

#### `packages/config`

Shared tooling configuration.

```text
packages/config/
├── eslint/
├── typescript/
├── prettier/
└── package.json
```

Useful for sharing:

- TypeScript config
- ESLint config
- Prettier config
- Jest/Vitest config

---

## Example for a TypeScript full-stack project

If you use **Next.js + NestJS + PostgreSQL**, a good structure would be:

```text
my-project/
├── apps/
│   ├── web/                 # Next.js frontend
│   │   ├── app/
│   │   ├── components/
│   │   ├── lib/
│   │   └── package.json
│   │
│   └── api/                 # NestJS backend
│       ├── src/
│       │   ├── modules/
│       │   ├── common/
│       │   ├── main.ts
│       │   └── app.module.ts
│       └── package.json
│
├── packages/
│   ├── types/               # Shared DTOs/types
│   ├── shared/              # Shared utilities
│   ├── ui/                  # Optional shared React components
│   └── config/              # Shared configs
│
├── docker-compose.yml
├── pnpm-workspace.yaml
├── turbo.json
├── package.json
└── README.md
```

---

## Root `package.json`

For a JavaScript/TypeScript monorepo, I usually recommend **pnpm workspaces** with **Turborepo**.

Example:

```json
{
  "name": "my-project",
  "private": true,
  "scripts": {
    "dev": "turbo dev",
    "build": "turbo build",
    "test": "turbo test",
    "lint": "turbo lint",
    "format": "prettier --write ."
  },
  "devDependencies": {
    "turbo": "latest",
    "typescript": "latest",
    "prettier": "latest"
  },
  "packageManager": "pnpm@9.0.0"
}
```

---

## `pnpm-workspace.yaml`

```yaml
packages:
  - "apps/*"
  - "packages/*"
```

---

## `turbo.json`

```json
{
  "tasks": {
    "dev": {
      "cache": false,
      "persistent": true
    },
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", ".next/**"]
    },
    "test": {
      "dependsOn": ["build"]
    },
    "lint": {}
  }
}
```

---

## Backend structure example

For the backend, prefer a feature/module-based structure:

```text
apps/api/src/
├── modules/
│   ├── auth/
│   │   ├── auth.controller.ts
│   │   ├── auth.service.ts
│   │   ├── auth.routes.ts
│   │   └── auth.types.ts
│   │
│   ├── users/
│   │   ├── users.controller.ts
│   │   ├── users.service.ts
│   │   ├── users.repository.ts
│   │   └── users.types.ts
│   │
│   └── products/
│       ├── products.controller.ts
│       ├── products.service.ts
│       └── products.repository.ts
│
├── common/
│   ├── middleware/
│   ├── errors/
│   ├── guards/
│   └── utils/
│
├── config/
├── database/
└── main.ts
```

This scales better than grouping by technical layer like:

```text
controllers/
services/
repositories/
models/
```

Feature-based grouping keeps related files close together.

---

## Frontend structure example

For frontend, also prefer feature-based organization:

```text
apps/web/src/
├── app/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── api.ts
│   │   └── types.ts
│   │
│   ├── users/
│   │   ├── components/
│   │   ├── hooks/
│   │   └── api.ts
│   │
│   └── products/
│       ├── components/
│       ├── hooks/
│       └── api.ts
│
├── components/
│   ├── layout/
│   └── common/
│
├── lib/
├── config/
└── styles/
```

Use:

- `features/` for business-specific UI and logic
- `components/common` for generic app-specific components
- `lib/` for app-level helpers
- `packages/ui` for reusable design-system components

---

## Environment variables

Avoid one huge `.env` shared by everything.

Use:

```text
my-project/
├── .env.example
├── apps/
│   ├── web/
│   │   ├── .env.local
│   │   └── .env.example
│   │
│   └── api/
│       ├── .env
│       └── .env.example
```

Example:

```text
apps/web/.env.local
NEXT_PUBLIC_API_URL=http://localhost:4000
```

```text
apps/api/.env
DATABASE_URL=postgresql://user:password@localhost:5432/mydb
JWT_SECRET=secret
PORT=4000
```

Never commit real `.env` files.

---

## Docker structure

For local development, keep Docker at the root:

```text
docker-compose.yml
```

Example services:

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

If you containerize frontend and backend:

```text
apps/
├── web/
│   └── Dockerfile
└── api/
    └── Dockerfile
```

---

## Best practices

### 1. Keep apps deployable independently

The frontend should not require backend source code directly, except shared packages like types or validation schemas.

Good:

```ts
import type { UserDto } from "@repo/types";
```

Avoid:

```ts
import { UserService } from "../../api/src/modules/users/users.service";
```

---

### 2. Share only stable code

Do not over-share everything.

Good things to share:

- types
- schemas
- constants
- UI components
- config
- utility functions

Avoid sharing:

- backend services
- database models directly
- app-specific business logic
- framework-specific code unless intentionally packaged

---

### 3. Use package boundaries

Each app/package should have its own `package.json`.

Example:

```json
{
  "name": "@repo/types",
  "version": "0.0.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts"
}
```

Then import like:

```ts
import type { UserDto } from "@repo/types";
```

---

### 4. Use TypeScript path aliases carefully

Prefer workspace package imports:

```ts
import { formatDate } from "@repo/shared";
```

Rather than deep relative imports:

```ts
import { formatDate } from "../../../packages/shared/src/date";
```

---

### 5. Keep database code in backend

A common mistake is putting database schema/models in a shared package and importing them everywhere.

Usually better:

```text
apps/api/src/database/
```

Then expose safe DTOs/types through:

```text
packages/types/
```

The frontend should not know about your database internals.

---

## Recommended tools

For JavaScript/TypeScript monorepos:

### Best default stack

```text
pnpm workspaces + Turborepo
```

Why:

- fast installs
- simple workspace management
- good caching
- easy task running
- works well with frontend/backend apps

Alternative:

```text
Nx
```

Use Nx if you want:

- more opinionated tooling
- code generators
- dependency graph
- larger enterprise-style structure

---

## My recommended final structure

For most projects, I would start with this:

```text
my-project/
├── apps/
│   ├── web/
│   └── api/
│
├── packages/
│   ├── types/
│   ├── shared/
│   └── config/
│
├── infra/
├── scripts/
├── docs/
│
├── docker-compose.yml
├── pnpm-workspace.yaml
├── turbo.json
├── tsconfig.base.json
├── package.json
├── .gitignore
├── .env.example
└── README.md
```

Add `packages/ui` only if you really need shared UI components.

## Short recommendation

If you want a clean, scalable default:

```text
apps/web      → frontend app
apps/api      → backend app
packages/types → shared API types/DTOs
packages/shared → shared utilities and validation
packages/config → shared TS/ESLint/Prettier config
infra/        → Docker, deployment, Terraform, Nginx
scripts/      → automation scripts
```

This structure is simple enough for small teams and scalable enough for large projects.