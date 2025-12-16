.PHONY: up down logs dev backend-dev frontend-dev

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

dev:
	make -j 2 backend-dev frontend-dev

backend-dev:
	cd backend && go run cmd/server/main.go

frontend-dev:
	cd frontend && npm run dev
