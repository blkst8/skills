# Auditing and Restructuring an Existing Repo

Restructuring differs from scaffolding: real code, real history, real risk.
The audit finds what's wrong; the migration moves things without breaking
them. Read both sections — audit first, even for a restructure request,
because the audit output *is* the migration plan.

## Audit checklist

Walk the repo and answer these. Each "no" is a finding; each finding maps
to a fix in the migration steps.

1. **Independent deployability** — does any app import another app's
   source (`../../api/src/...`)? Grep for cross-app relative imports.
2. **Sharing discipline** — what's in shared code? Backend services,
   database models, or app-specific logic in a package are over-sharing.
   Duplicated types across apps are under-sharing.
3. **Package boundaries** — do imports go through package names
   (`@repo/...`) or deep relative paths (`../../../packages/...`)?
4. **Database isolation** — where do schema/models live? In a shared
   package, or (correctly) in the backend?
5. **Feature grouping** — are the apps organized by feature/module or by
   technical layer (`controllers/`, `services/` flat dirs)?
6. **Environment hygiene** — one giant `.env`? Real env files committed?
7. **Tool wiring** — does `install` → `build` → `test` actually pass at
   the root, or does the repo have ritual config that's never run?

Present findings as a short list: what's wrong, why it hurts (tie to the
five goals), what to do. Severity-order them — import-boundary violations
are urgent; layer-grouping is cosmetic.

## Migration: restructure step by step

The safe order moves structure first, wiring second, imports last. Each
step leaves the repo in a working state if the previous one was verified.

1. **Inventory** — list every top-level artifact and where it *should* live
   per `layout.md`. Produce a move map before touching anything:
   `client/ → apps/web`, `server/ → apps/api`, duplicated types →
   `packages/types`, shared utils → `packages/shared`.
2. **Create the skeleton** — root files (`pnpm-workspace.yaml`,
   `turbo.json`, `tsconfig.base.json`, root `package.json`), empty
   `apps/` and `packages/` with their `package.json`s.
3. **Move with `git mv`** — never copy-and-delete: `git mv` preserves
   history and makes the restructure reviewable as a rename, not a
   delete+add. Do the big moves in one commit if possible.
4. **Extract packages** — for duplicated code, pick the better copy,
   move it to the package, and re-point *both* consumers at it. For code
   shared via deep relative imports, move once, import via `@repo/*`.
5. **Rewrite imports** — mechanical after the moves: relative → workspace
   package. Run typecheck after each app's rewrite.
6. **Split the environment** — carve the giant `.env` into per-app files
   with `.env.example`s; update `.gitignore`; remove committed secrets
   from tracking (`git rm --cached`).
7. **Verify at each seam** — `pnpm install`, then `pnpm build` (dependency
   order works), then `pnpm test`, then run each app independently. A
   migration isn't done until both apps start alone.

## Practical mechanics you'll hit mid-migration

These come up in almost every real restructure — handle them deliberately
instead of improvising:

- **CommonJS backend consuming a TS types package.** A plain-JS Express
  server can't `require("@repo/types")` at runtime (it's types-only). Don't
  convert the module system just to share types — consume them via JSDoc:

  ```js
  /** @param {import("@repo/types").SignupRequestDto} body */
  ```

  or type files where the toolchain supports it. Module-system conversion
  is a separate, explicit change — never a side effect of the restructure.
- **Environment loading.** Per-app env files need a loader in the app: for
  Node ≥ 20.6, `node --env-file=.env` does it with zero dependencies —
  prefer that over adding dotenv.
- **Dependencies that the move exposes.** Splitting the original
  package.json re-assigns every dep to its real owner. When a line lands
  nowhere obviously (a stated-but-unused zod, or a vite plugin the config
  imports but the manifest omitted), add it to the owning package and flag
  it to the user rather than silently dropping it.
- **Handler bodies stay byte-identical.** The only acceptable changes when
  moving: import paths, type references, and comments. Everything else
  waits — this is what makes the restructure diff reviewable.

### Moving a JS project into a polyglot repo

If the backend is Go/Python, don't fabricate a `package.json` for it during
migration — read `references/polyglot.md` first and choose the orchestration
story (Makefile or thin wrapper) before moving files.

## Judging scope: full restructure vs. incremental

Big-bang is right when the repo is small or pre-release. For a large,
active repo, propose an incremental path: fix the import boundaries and
package extraction first (they cause the real pain), leave cosmetic moves
(layer grouping) for later PRs. Say which one you're doing and why.

## What not to do

- Don't reformat/restyle code during a restructure — move files and
  rewrite imports, nothing else. Mixed changes make review impossible.
- Don't invent packages with a single consumer "for symmetry".
- Don't move the database schema out of the backend into a package to make
  types shareable — that's the audit finding, not the fix; expose DTOs.
- Don't delete anything that looks dead without flagging it to the user
  first — in a migration, "unused" often means "used in production."
