.PHONY: up down logs dev start setup seed seed-force migrate-up migrate-down migrate-status migrate-create migrate-docker-up migrate-docker-down migrate-docker-status cleanup-audit-logs security security-gosec security-vulncheck security-lint test test-backend test-frontend test-e2e test-coverage

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
	docker compose exec backend go run cmd/seed/main.go

seed-force:
	docker compose exec -e FORCE_SEED=true backend go run cmd/seed/main.go

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

# Maintenance
cleanup-audit-logs:
	./scripts/cleanup-audit-logs.sh

# Security Scanning
security: security-gosec security-vulncheck security-lint
	@echo "All security checks completed!"

security-gosec:
	@echo "Running gosec security scanner..."
	cd backend && gosec -fmt=json -out=gosec-report.json -stdout -verbose=text ./...

security-vulncheck:
	@echo "Running govulncheck for known vulnerabilities..."
	cd backend && govulncheck ./...

security-lint:
	@echo "Running golangci-lint with security checks..."
	cd backend && golangci-lint run --config=../.golangci.yml ./...

# Testing
test: test-backend test-frontend
	@echo "All tests completed!"

test-backend:
	@echo "Running backend tests..."
	cd backend && go test -v -race ./...

test-frontend:
	@echo "Running frontend unit tests..."
	cd frontend && npm run test:run

test-e2e:
	@echo "Running E2E tests..."
	cd frontend && npm run test:e2e

test-coverage:
	@echo "Running tests with coverage..."
	cd backend && go test -v -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
	cd frontend && npm run test:coverage
