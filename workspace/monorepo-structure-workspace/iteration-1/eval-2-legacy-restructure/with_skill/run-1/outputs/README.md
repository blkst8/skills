# legacy-app

User signup demo, restructured as a pnpm + Turborepo monorepo.

## Structure

```
apps/
├── web/            # React + Vite frontend
└── api/            # Express + pg backend
packages/
├── types/          # shared DTOs (UserDto, SignupRequest) — used by both apps
└── shared/         # shared runtime helpers (date formatting)
```

## Run

Requires pnpm. Start local infrastructure first (optional):

```
docker compose up -d        # postgres + redis
```

Then:

```
pnpm install
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env.local
pnpm dev                    # runs api + web via turbo
```

Or run each app independently:

```
pnpm --filter api dev       # http://localhost:4000
pnpm --filter web dev       # http://localhost:5173
```
