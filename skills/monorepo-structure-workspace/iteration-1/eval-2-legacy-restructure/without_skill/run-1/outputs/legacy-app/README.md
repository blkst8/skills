# legacy-app

User signup demo, restructured as a pnpm + Turborepo monorepo.

## Structure

- `apps/web` — React client (Vite)
- `apps/api` — Express API; routes, models, db, and crypto are split into
  separate layers under `src/`
- `packages/shared` — types shared by web and api (`UserDto`,
  `SignupRequest`) plus date helpers (`formatDate`, `relativeTime`)

## Run

```
pnpm install
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env
pnpm dev        # turbo runs api + web together
```

Server: http://localhost:4000 — Client: http://localhost:5173

The old single root `.env` is now split per app: database/JWT/Redis config
lives in `apps/api/.env`, the frontend API URL in `apps/web/.env` (Vite only
exposes `VITE_`-prefixed vars to the browser). `.env` files are gitignored;
`.env.example` documents the required variables.
