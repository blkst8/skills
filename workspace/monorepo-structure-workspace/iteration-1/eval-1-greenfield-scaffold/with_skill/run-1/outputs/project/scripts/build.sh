#!/usr/bin/env bash
# Build every workspace package and app in dependency order.
set -euo pipefail
cd "$(dirname "$0")/.."
pnpm build
