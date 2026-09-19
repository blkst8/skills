# TypeScript Stack: pnpm workspaces + Turborepo

Why this default: fast installs, simple workspace management, good
caching, one-command task running across apps. Nx is the alternative when
generators and dependency graphs are wanted; pnpm + Turbo is the lighter
default.

## Root package.json

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
    "turbo": "^2",
    "typescript": "^5",
    "prettier": "^3"
  },
  "packageManager": "pnpm@9.0.0"
}
```

The root is an orchestrator, not an app — no runtime dependencies belong
here. Scripts delegate to Turbo, which fans them out to whatever packages
define the task.

Pin a real version in `packageManager`; the exact pnpm version matters for
lockfile compatibility across machines.

## pnpm-workspace.yaml

```yaml
packages:
  - "apps/*"
  - "packages/*"
```

## Declaring internal dependencies

An app that needs shared code declares a normal-looking dependency:

```json
{
  "dependencies": {
    "@repo/types": "workspace:*"
  }
}
```

`workspace:*` resolves to the local package, is what gets pnpm linking
instead of installing, and behaves correctly when publishing later. After
adding or changing workspace dependencies, run `pnpm install` once at the
root to refresh links.

## turbo.json

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

Reading it:

- `dependsOn: ["^build"]` — build dependencies' *packages* first
  (`^` = topological). `@repo/shared` builds before `apps/web` does.
- `"test": { "dependsOn": ["build"] }` (no `^`) — run the *same package's*
  build before its tests.
- `outputs` lists what caching should fingerprint. `dist/**` and `.next/**`
  cover most stacks; add the app's actual output dirs so cache hits work.
- `dev` is `persistent` (long-running) and never cached.

Turbo only sees packages that have a `package.json` with the matching
script — an app without a `test` script is simply skipped, no error.

## tsconfig.base.json

```json
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "preserve",
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

Apps extend it:

```json
{
  "extends": "@repo/config/typescript/base.json",
  "include": ["src"]
}
```

Prefer workspace package imports over path aliases for cross-package code:

```ts
import { formatDate } from "@repo/shared";        // good
import { formatDate } from "../../../packages/shared/src/date"; // never
```

Deep relative imports break silently on file moves, bypass the package
boundary, and hide dependency direction. Use `paths` aliases only inside a
single app (`@/` → `apps/web/src/`), never as a substitute for packages.

## Shared ESLint config

With `packages/config/eslint/`, the root config is tiny:

```js
// eslint.config.js (flat config)
export { default } from "@repo/config/eslint";
```

Individual apps extend and add framework rules (Next.js, NestJS) on top.
One place to tighten rules for the whole repo.

## Verify the scaffold

A scaffold isn't done until the wiring is proven:

1. `pnpm install` at root succeeds
2. `pnpm build` builds packages before apps (order visible in output)
3. `apps/web` imports something from `@repo/types` and type-checks
4. `pnpm test` / `pnpm lint` run across all packages without error

If any step fails, fix the config — don't hand the user a structure that
looks right but doesn't run.
