# Where to put the shared Zod schema and the Prisma models

## 1. The Zod signup schema → a new validation package: `packages/validation`

With exactly one frontend and one backend, it's tempting to just put the schema in `packages/types`, but resist that: your `packages/types` should stay pure types (zero runtime code, tree-shakeable, no dependencies). A Zod schema is runtime validation logic with a `zod` dependency — different animal. Mixing runtime code into a types package makes every consumer pull in `zod` and blurs the package's contract.

Since both `apps/web` and `apps/api` genuinely need the same schema (form validation client-side, request-body validation server-side), this is a real case for a new shared package:

```
packages/
  types/          # existing — keep it types-only
  validation/     # NEW
    package.json  # name: "@your-org/validation", deps: zod
    src/
      index.ts    # re-exports
      auth.ts     # signupSchema (and later loginSchema, etc.)
    tsconfig.json
```

Suggested structure:

```ts
// packages/validation/src/auth.ts
import { z } from "zod";

export const signupSchema = z.object({
  email: z.string().email(),
  password: z
    .string()
    .min(8)
    .regex(/[A-Z]/, "Must contain an uppercase letter")
    .regex(/[0-9]/, "Must contain a digit"),
});

export type SignupInput = z.infer<typeof signupSchema>;
```

```json
// packages/validation/package.json
{
  "name": "@your-org/validation",
  "version": "0.0.0",
  "private": true,
  "main": "src/index.ts",
  "types": "src/index.ts",
  "dependencies": { "zod": "^3.x" },
  "devDependencies": { "typescript": "^5.x" }
}
```

Then:

- `apps/web/package.json` → `"@your-org/validation": "workspace:*"`
- `apps/api/package.json` → `"@your-org/validation": "workspace:*"`

In the Fastify handler:

```ts
import { signupSchema } from "@your-org/validation";

app.post("/signup", async (req, reply) => {
  const parsed = signupSchema.safeParse(req.body);
  if (!parsed.success) {
    return reply.code(400).send({ error: parsed.error.flatten() });
  }
  // ...
});
```

And in the React form, use the same schema for client-side validation so error messages match exactly. One schema, one source of truth for the contract.

A note on `main: "src/index.ts"`: pointing `main`/`types` directly at the TypeScript source works fine in a pnpm monorepo where both apps bundle with Vite/esbuild/tsup and compile through the same TS version. If your api is deployed standalone or you hit friction, add a build step (tsup) and point at `dist/` instead — but don't bother preemptively.

## 2. The Prisma models → `packages/db`, NOT the same package

Moving the Prisma schema/client into a shared package is a good idea — the web app can then use generated types (e.g. `User`) without importing from `apps/api`. But it should be its own dedicated package, `packages/db`, separate from the validation package:

- `packages/db` carries the Prisma schema, generated client, and the migration history. Its concern is persistence.
- `packages/validation` is pure, dependency-light contract logic. Anyone consuming validation shouldn't transitively engage a database client.

(You could argue for one `packages/shared` grabbing everything, but don't — with two distinct concerns already, a grab-bag shared package rots fast. Small, purpose-named packages are cheaper in a workspace monorepo.)

```
packages/db/
  package.json        # "@your-org/db", deps: prisma + @prisma/client
  prisma/
    schema.prisma     # moved from apps/api/src
    migrations/       # move these too!
  src/
    index.ts          # export a singleton prisma client
    generated/        # prisma generate output (gitignored or committed, your call)
```

```ts
// packages/db/src/index.ts
import { PrismaClient } from "@your-org/db/generated"; // or "@prisma/client" if output goes to node_modules
import { singleton } from "./utils/singleton";

export const prisma = singleton("prisma", () => new PrismaClient());
```

Caveats your teammate should know:

1. **Only move the schema and migrations, not the client instantiation scattered through the api.** Replace `apps/api` imports of the prisma client with `@your-org/db`.
2. **A word of caution for the web app**: importing *types* from `@your-org/db` in `apps/web` is fine and is the real win. Importing the *PrismaClient runtime* into a browser bundle is not fine — the web app should never query the database directly; it should call the api. Keep that boundary; sharing the schema is for shared types, not shared database access.
3. Update `schema.prisna`'s `generator` output path if needed, and run `prisma generate` via the db package's `postinstall` or a root script.
4. Env vars (`DATABASE_URL`) become a concern of wherever the prisma client is instantiated — with the singleton in `packages/db`, both apps inherit it via `.env` in each app or at the root.

## Summary

| Concern | Destination | Why |
|---|---|---|
| Zod signup schema | `packages/validation` (new) | Runtime validation shared by web + api; keep `packages/types` types-only |
| Prisma models | `packages/db` (new, dedicated) | Persistence concern isolated; web imports types only, never the client |

Both are created via `pnpm` workspace globs (`packages/*` in root `pnpm-workspace.yaml` if not already) and wired with `"workspace:*"` versions. Don't create a catch-all `packages/shared`, and don't add a `packages/ui` — with exactly one frontend app, that would be premature; UI components stay in `apps/web` until a second consumer exists.
