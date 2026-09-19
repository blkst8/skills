# Convenience entrypoints so you don't have to remember both toolchains.
# Go side never needs a package.json — it's driven by go.mod.

.PHONY: dev-web dev-api build test

dev-web:
	pnpm --filter @sideproject/web dev

dev-api:
	go run ./apps/api

build:
	pnpm --filter @sideproject/web build
	go build -o apps/api/bin/api ./apps/api

test:
	pnpm -r test
	go test ./apps/api/...

install:
	pnpm install
