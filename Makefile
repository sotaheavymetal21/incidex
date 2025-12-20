.PHONY: up down logs ps restart \
	dev backend-dev frontend-dev \
	backend-build backend-test backend-fmt \
	frontend-build frontend-start frontend-lint \
	setup setup-backend setup-frontend \
	seed seed-docker

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

ps:
	docker-compose ps

restart: down up

dev:
	make -j 2 backend-dev frontend-dev

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
	cd backend && go run cmd/seed/main.go

seed-docker:
	docker-compose exec backend go run cmd/seed/main.go
