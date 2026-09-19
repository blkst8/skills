# Cross-language task orchestration. pnpm only knows about JS workspaces,
# so this Makefile is the one canonical place to run the whole repo.
# JS-only commands still work directly: `pnpm --filter web dev`.

.DEFAULT_GOAL := help

.PHONY: help dev web api build test generate db-up db-down

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

dev: ## Run web + api together
	$(MAKE) -j2 web api

web: ## Run the web dev server (Vite)
	pnpm --filter web dev

api: ## Run the Go API server
	go run ./apps/api/cmd/api

build: ## Build all apps
	pnpm --filter web build
	cd apps/api && go build -o bin/api ./cmd/api

test: ## Run Go tests + web typecheck
	cd apps/api && go test ./...
	pnpm --filter web typecheck

generate: ## Regenerate TS types from the OpenAPI contract
	pnpm --filter api-schema generate

db-up: ## Start local Postgres via docker compose
	docker compose up -d

db-down: ## Stop local infrastructure
	docker compose down
