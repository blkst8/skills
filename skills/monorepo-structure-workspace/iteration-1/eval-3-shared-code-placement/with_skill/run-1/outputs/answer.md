# Where should these go?

Short answers: **the zod signup schema → `packages/shared`** (a new shared package), and **the Prisma models stay in `apps/api`** — don't move them into a shared package. Details below.

## 1. The zod signup schema → `packages/shared`

A runtime validation schema is exactly the kind of thing `packages/` exists for: stable, pure logic, needed by two apps, no deployment concerns attached. It's shared *code*, not just shared *types* — so it does **not** belong in your existing `packages/types`, which should stay limited to types/DTOs. Create a dedicated package:

```
packages/shared/
├── package.json        # name: "@repo/shared", main/exports -> src/index.ts
├── tsconfig.json
└── src/
    ├── index.ts        # export * from "./auth/signup-schema"
    └── auth/
        └── signup-schema.ts   # the zod schema
```

Then both consumers import it as a workspace dependency:

```ts
// apps/web/src/features/signup/SignupForm.tsx
import { signupSchema } from "@repo/shared";

// apps/api/src/modules/auth/auth.controller.ts
import { signupSchema } from "@repo/shared";
```

Add `"@repo/shared": "workspace:*"` to the `dependencies` of both `apps/web/package.json` and `apps/api/package.json`, and export the derived types from it too (`export type SignupInput = z.infer<typeof signupSchema>`) so both sides agree on the shape without duplicating anything. Note that `zod` should be a regular dependency of `packages/shared` itself, not just of the apps — the package owns what it needs at runtime.

This also replaces any duplicated email/password-strength logic on either side: the web form validates client-side with the same schema the API uses server-side, so the rules can never drift apart.

## 2. The Prisma models stay in `apps/api` — don't share them

I'd push back on your teammate's suggestion. Prisma models are database code, and database code belongs in the backend (`apps/api/src/database/` in a clean layout). If `apps/web` imports them directly, three bad things happen:

- **You drag the ORM into the frontend.** Importing the Prisma client means the web app's build now depends on Prisma's generated client, engine binaries, and toolchain — the frontend is no longer deployable without the backend's database machinery.
- **Every schema change breaks both apps at once.** The moment you add a column or change a relation, the web app's build fails too. The database becomes unable to evolve without a frontend rebuild/release.
- **The frontend learns the world through database internals.** Your DB rows will contain fields the signup form and UI should never see (password hashes, internal flags, timestamps you don't render). Sharing raw models leaks all of that by default.

The pattern that gets your teammate what they actually want — the web app importing "user types" directly — is **DTOs in the existing `packages/types`**:

```ts
// packages/types/src/auth.ts
export interface SignupRequestDto {
  email: string;
  password: string;
}

export interface UserDto {
  id: string;
  email: string;
  name: string;
  createdAt: string;
}
```

The API maps Prisma models → DTOs in its responses (`UserDto` never has `passwordHash`), and the web app does:

```ts
import type { UserDto } from "@repo/types";
```

That import works everywhere, survives refactors, keeps the dependency direction explicit, and lets the database schema change without touching the web app as long as the DTO contract holds.

## Rule of thumb for future calls like this

Ask: **do both apps need it, and is it stable, and does it carry no backend baggage?** Types/DTOs, validation schemas, constants, and pure utilities → `packages/`. Anything wired to a database, ORM, service layer, or one app's business logic → stays in the app that owns it. Also: since you have exactly one frontend app and no design system, resist creating `packages/ui` — that's another premature-sharing trap until a second consumer exists.

## Summary

| Item | Destination | Why |
|---|---|---|
| zod signup schema (runtime code) | `packages/shared` (new) | Stable, pure, used by both apps; runtime code ≠ types package |
| Prisma models | stay in `apps/api` | ORM + DB concerns must not leak into the web app |
| What the web actually needs from Prisma | DTOs in existing `packages/types` | Explicit contract, no ORM dependency, schema can evolve independently |
