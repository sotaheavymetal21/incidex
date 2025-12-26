.PHONY: up down logs ps restart \
	dev backend-dev frontend-dev \
	local-dev local-up local-down local-logs local-setup \
	backend-build backend-test backend-fmt \
	frontend-build frontend-start frontend-lint \
	setup setup-backend setup-frontend \
	seed seed-docker docker-build docker-rebuild

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

logs-backend:
	docker compose logs -f backend

ps:
	docker compose ps

restart: down up

docker-build:
	docker compose build

docker-rebuild:
	docker compose down
	docker compose build --no-cache
	docker compose up -d

dev:
	make -j 2 backend-dev frontend-dev

# ローカル開発環境の起動（インフラ + フロントエンド + バックエンド）
local-dev: local-up
	@echo "Starting local development environment..."
	@echo "Infrastructure services are running in Docker"
	@echo "Starting backend and frontend..."
	@make -j 2 backend-dev frontend-dev

# インフラサービスのみ起動（DB, Redis, MinIO）
local-up:
	@echo "Starting infrastructure services (PostgreSQL, Redis, MinIO)..."
	@docker compose up -d db redis minio
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Infrastructure services are ready!"

# インフラサービスの停止
local-down:
	@echo "Stopping infrastructure services..."
	@docker compose stop db redis minio

# インフラサービスのログ確認
local-logs:
	@docker compose logs -f db redis minio

# 初回セットアップ（依存関係インストール + インフラ起動 + シードデータ投入）
local-setup: setup local-up
	@echo "Setup completed!"
	@echo "Run 'make seed' to seed the database with test data"

backend-dev:
	cd backend && go run cmd/server/main.go

frontend-dev:
	cd frontend && npm run dev

backend-build:
	cd backend && go build ./...

backend-test:
	cd backend && go test ./...

backend-fmt:
	cd backend && gofmt ./...

frontend-build:
	cd frontend && npm run build

frontend-start:
	cd frontend && npm run start

frontend-lint:
	cd frontend && npm run lint

setup: setup-backend setup-frontend

setup-backend:
	cd backend && go mod tidy

setup-frontend:
	cd frontend && npm install

seed:
	@echo "Generating test data..."
	@node scripts/generate-test-data.js

seed-docker:
	docker compose exec backend ./seeder
