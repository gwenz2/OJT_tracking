# Convenience wrappers around backend/frontend tooling.
# Windows users: run `make` via Git Bash, WSL, or a `make` port (e.g. chocolatey `make`).

.PHONY: help dev backend frontend migrate-up migrate-down seed-admin seed-demo test lint fmt vet install compose-up compose-down

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36}%-18s\033[0m %s\n", $$1, $$2}'

dev: ## Run backend and frontend together (two terminals)
	@echo "Open two terminals: 'make backend' and 'make frontend'"

backend: ## Run Go backend (migrations applied automatically)
	cd backend && go run ./cmd/api

frontend: ## Run Vite dev server (proxies /api -> :8080)
	cd frontend && npm run dev

migrate-up: ## Apply database migrations
	cd backend && go run ./cmd/api --migrate-up

migrate-down: ## Roll back the latest migration (dev only)
	cd backend && go run ./cmd/api --migrate-down

seed-admin: ## Create initial admin: ADMIN_EMAIL=.. ADMIN_PASSWORD=.. make seed-admin
	cd backend && go run ./cmd/api --seed-admin

seed-demo: ## Create demo trainee/site/assignment (dev only)
	cd backend && go run ./cmd/seed

test: ## Run backend tests (integration/contract need Postgres+MinIO via compose-up)
	cd backend && go test ./internal/... ./tests/...
	cd frontend && npm test -- --run

test-race: ## Backend tests with race detector (Phase 5 gate)
	cd backend && go test -race -count=3 ./internal/... ./tests/...

lint: ## Lint backend and frontend
	cd backend && go vet ./...
	cd frontend && npm run lint

fmt: ## Format Go code
	cd backend && go fmt ./...

vet: ## Run go vet
	cd backend && go vet ./...

compose-up: ## Start local PostgreSQL + MinIO
	docker compose up -d

compose-down: ## Stop local dependencies
	docker compose down

install: ## Install backend and frontend dependencies
	cd backend && go mod download
	cd frontend && npm ci
