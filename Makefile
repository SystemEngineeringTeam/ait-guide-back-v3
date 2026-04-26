.PHONY: help dev dev-build prod prod-build down down-v \
       logs logs-db \
       run build test test-cover fmt vet lint \
       db-up db-psql

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ─── Docker (dev) ────────────────────────────────────────
dev: ## Start dev environment (db + app with hot-reload)
	docker compose up -d db app

dev-build: ## Build and start dev environment
	docker compose up -d --build db app

COMPOSE_PROD = docker compose -f compose.yaml -f compose.prod.yaml

# ─── Docker (prod) ───────────────────────────────────────
prod: ## Start production environment (db + app)
	$(COMPOSE_PROD) up -d

prod-build: ## Build and start production environment
	$(COMPOSE_PROD) up -d --build

# ─── Docker (common) ────────────────────────────────────
down: ## Stop all containers
	docker compose down

down-v: ## Stop all containers and remove volumes
	docker compose down -v

logs: ## Tail app logs
	docker compose logs -f app

logs-db: ## Tail db logs
	docker compose logs -f db

# ─── Database ────────────────────────────────────────────
db-up: ## Start only the database
	docker compose up -d db

db-psql: ## Connect to database via psql
	docker compose exec db psql -U aitguide -d aitguide

# ─── Local development ──────────────────────────────────
run: ## Run server locally
	go run cmd/server/main.go

build: ## Build binary
	go build -o bin/server cmd/server/main.go

test: ## Run tests
	go test ./...

test-cover: ## Run tests with coverage
	go test -cover ./...

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run
