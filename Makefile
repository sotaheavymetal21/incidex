.PHONY: up down logs dev start setup seed seed-force migrate-up migrate-down migrate-status migrate-create migrate-docker-up migrate-docker-down migrate-docker-status

# Docker
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# Development
dev:
	docker compose up -d --build
	cd frontend && npm run dev

# Start: Setup + Migrate + Seed + Start all services
start:
	@echo "Setting up dependencies..."
	cd backend && go mod tidy
	cd frontend && npm install
	@echo "Building and starting containers..."
	docker compose up -d --build
	@echo "Running migrations..."
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" up
	@echo "Seeding database..."
	cd backend && go run cmd/seed/main.go
	@echo "Starting frontend..."
	cd frontend && npm run dev

# Setup
setup:
	cd backend && go mod tidy
	cd frontend && npm install

# Database
seed:
	cd backend && go run cmd/seed/main.go

seed-force:
	cd backend && FORCE_SEED=true go run cmd/seed/main.go

# Migration commands
MIGRATE_DB_URL ?= postgres://user:password@localhost:5432/incidex?sslmode=disable

migrate-up:
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" down

migrate-status:
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" status

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please specify a migration name with 'make migrate-create name=your_migration_name'"; \
		exit 1; \
	fi
	cd backend && goose -dir migrations create $(name) sql

# Docker migration commands
migrate-docker-up:
	docker compose exec backend goose -dir /app/migrations postgres "postgres://user:password@postgres:5432/incidex?sslmode=disable" up

migrate-docker-down:
	docker compose exec backend goose -dir /app/migrations postgres "postgres://user:password@postgres:5432/incidex?sslmode=disable" down

migrate-docker-status:
	docker compose exec backend goose -dir /app/migrations postgres "postgres://user:password@postgres:5432/incidex?sslmode=disable" status
