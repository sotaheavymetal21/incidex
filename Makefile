.PHONY: up down logs ps restart \
	dev backend-dev frontend-dev \
	local-dev local-up local-down local-logs local-setup \
	backend-build backend-test backend-fmt \
	frontend-build frontend-start frontend-lint \
	setup setup-backend setup-frontend \
	seed seed-force seed-docker docker-build docker-rebuild \
	migrate-up migrate-down migrate-status migrate-create migrate-reset migrate-docker-up migrate-docker-down migrate-docker-status

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

# ローカル開発環境の起動（インフラ + Backendコンテナ + フロントエンド + バックエンド）
local-dev: local-up
	@echo "Starting local development environment..."
	@echo "Infrastructure services (including backend container) are running in Docker"
	@echo "Starting backend and frontend..."
	@make -j 2 backend-dev frontend-dev

# インフラサービスのみ起動（DB, Redis, MinIO, Backend）
local-up:
	@echo "Starting infrastructure services (PostgreSQL, Redis, MinIO, Backend)..."
	@docker compose up -d db redis minio backend
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Infrastructure services are ready!"

# インフラサービスの停止
local-down:
	@echo "Stopping infrastructure services..."
	@docker compose stop db redis minio backend

# インフラサービスのログ確認
local-logs:
	@docker compose logs -f db redis minio backend

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

# ローカルでGoのシードコマンドを実行（データベースに直接接続）
# TEST_USER_PASSWORD環境変数が設定されていない場合は、デフォルトパスワード（admin1234）が使用されます
seed:
	@echo "Seeding database with test data using Go seeder..."
	@cd backend && go run cmd/seed/main.go

# データベースをクリアしてからシードを実行（既存データを削除）
seed-reset:
	@echo "WARNING: This will delete all existing data!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "Clearing database tables..."; \
		cd backend && go run -tags reset cmd/seed/main.go || \
		psql postgres://user:password@localhost:5432/incidex?sslmode=disable -c "TRUNCATE TABLE users, tags, incidents, incident_activities, attachments, notification_settings, incident_templates, post_mortems, action_items, audit_logs CASCADE;" || \
		echo "Please manually clear the database tables or use: make seed-force"; \
		make seed; \
	else \
		echo "Seed reset cancelled"; \
	fi

# 強制的にシードを実行（既存データを無視）
seed-force:
	@echo "Force seeding database (will skip existing data check)..."
	@cd backend && FORCE_SEED=true go run cmd/seed/main.go

# Dockerコンテナ内でシードコマンドを実行
# TEST_USER_PASSWORD環境変数が設定されていない場合は、ランダムなパスワードが自動生成されます
seed-docker:
	@echo "Seeding database using Docker container..."
	@docker compose exec backend ./seeder

# Database Migration Commands (Local)
# DATABASE_URLが設定されていない場合はデフォルト値を使用（.envファイルの設定に合わせる）
MIGRATE_DB_URL ?= postgres://user:password@localhost:5432/incidex?sslmode=disable

migrate-up:
	@echo "Running database migrations..."
	@cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" up

migrate-down:
	@echo "Rolling back last migration..."
	@cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" down

migrate-status:
	@echo "Checking migration status..."
	@cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" status

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "ERROR: Migration name is required"; \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)..."
	@cd backend && goose -dir migrations create $(name) sql

migrate-reset:
	@echo "WARNING: This will reset ALL migrations!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" reset; \
	else \
		echo "Migration reset cancelled"; \
	fi

# Database Migration Commands (Docker)
migrate-docker-up:
	@echo "Running database migrations in Docker..."
	@docker compose exec backend goose -dir migrations postgres "postgres://user:password@db:5432/incidex?sslmode=disable" up

migrate-docker-down:
	@echo "Rolling back last migration in Docker..."
	@docker compose exec backend goose -dir migrations postgres "postgres://user:password@db:5432/incidex?sslmode=disable" down

migrate-docker-status:
	@echo "Checking migration status in Docker..."
	@docker compose exec backend goose -dir migrations postgres "postgres://user:password@db:5432/incidex?sslmode=disable" status
