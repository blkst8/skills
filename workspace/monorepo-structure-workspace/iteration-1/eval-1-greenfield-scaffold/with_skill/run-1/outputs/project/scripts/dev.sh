#!/usr/bin/env bash
# Bring up local infrastructure and start both apps in dev mode.
set -euo pipefail
cd "$(dirname "$0")/.."

# First run: create the per-app env files from their examples.
[ -f apps/api/.env ] || cp apps/api/.env.example apps/api/.env
[ -f apps/web/.env.local ] || cp apps/web/.env.example apps/web/.env.local

docker compose up -d postgres
pnpm dev
