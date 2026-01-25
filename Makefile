.PHONY: up down logs dev start setup seed seed-force migrate-up migrate-down migrate-status migrate-create migrate-docker-up migrate-docker-down migrate-docker-status cleanup-audit-logs security security-gosec security-vulncheck security-lint test test-backend test-frontend test-e2e test-coverage

# Docker コマンド
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# 開発環境
dev:
	docker compose up -d --build
	cd frontend && npm run dev

# 起動: セットアップ + マイグレーション + シード + 全サービス起動
start:
	@echo "依存関係をセットアップしています..."
	cd backend && go mod tidy
	cd frontend && npm install
	@echo "コンテナをビルドして起動しています..."
	docker compose up -d --build
	@echo "マイグレーションを実行しています..."
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" up
	@echo "データベースにシードデータを投入しています..."
	cd backend && go run cmd/seed/main.go
	@echo "フロントエンドを起動しています..."
	cd frontend && npm run dev

# セットアップ
setup:
	cd backend && go mod tidy
	cd frontend && npm install

# データベース
seed:
	docker compose exec backend go run cmd/seed/main.go

seed-force:
	docker compose exec -e FORCE_SEED=true backend go run cmd/seed/main.go

# マイグレーションコマンド
MIGRATE_DB_URL ?= postgres://user:password@localhost:5432/incidex?sslmode=disable

migrate-up:
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" down

migrate-status:
	cd backend && goose -dir migrations postgres "$(MIGRATE_DB_URL)" status

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "エラー: 'make migrate-create name=マイグレーション名' の形式でマイグレーション名を指定してください"; \
		exit 1; \
	fi
	cd backend && goose -dir migrations create $(name) sql

# Docker 環境用マイグレーションコマンド
migrate-docker-up:
	docker compose exec backend goose -dir /app/migrations postgres "postgres://user:password@postgres:5432/incidex?sslmode=disable" up

migrate-docker-down:
	docker compose exec backend goose -dir /app/migrations postgres "postgres://user:password@postgres:5432/incidex?sslmode=disable" down

migrate-docker-status:
	docker compose exec backend goose -dir /app/migrations postgres "postgres://user:password@postgres:5432/incidex?sslmode=disable" status

# メンテナンス
cleanup-audit-logs:
	./scripts/cleanup-audit-logs.sh

# セキュリティスキャン
security: security-gosec security-vulncheck security-lint
	@echo "すべてのセキュリティチェックが完了しました！"

security-gosec:
	@echo "gosec セキュリティスキャナーを実行しています..."
	cd backend && gosec -fmt=json -out=gosec-report.json -stdout -verbose=text ./...

security-vulncheck:
	@echo "govulncheck で既知の脆弱性をチェックしています..."
	cd backend && govulncheck ./...

security-lint:
	@echo "golangci-lint でセキュリティチェックを実行しています..."
	cd backend && golangci-lint run --config=../.golangci.yml ./...

# テスト
test: test-backend test-frontend
	@echo "すべてのテストが完了しました！"

test-backend:
	@echo "バックエンドテストを実行しています..."
	cd backend && go test -v -race ./...

test-frontend:
	@echo "フロントエンドユニットテストを実行しています..."
	cd frontend && npm run test:run

test-e2e:
	@echo "E2E テストを実行しています..."
	cd frontend && npm run test:e2e

test-coverage:
	@echo "カバレッジ付きでテストを実行しています..."
	cd backend && go test -v -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
	cd frontend && npm run test:coverage
