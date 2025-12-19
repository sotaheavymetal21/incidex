# CLAUDE.md

必ず日本語で回答してください。
Be sure to answer in Japanese.

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Incidex is a modern incident management system built with Go (backend) and Next.js (frontend), following Clean Architecture principles. The system provides JWT-based authentication, tag management for organizing incidents, and is designed for future incident tracking and timeline features.

## Development Commands

### Docker Environment (Recommended)
```bash
make up          # Start all services (PostgreSQL, Redis, MinIO, Backend, Frontend)
make down        # Stop and remove all containers
make logs        # View logs from all services
```

### Local Development
```bash
make dev         # Run both backend and frontend locally (parallel)
```

**Backend** (from `backend/` directory):
```bash
go run cmd/server/main.go   # Start backend server on port 8080
```

**Frontend** (from `frontend/` directory):
```bash
npm run dev      # Start Next.js dev server on port 3000
npm run build    # Build production bundle
npm run lint     # Run ESLint
```

### Testing
Currently no test files exist in the codebase. When adding tests:
- Go tests: Use standard `go test` commands
- Frontend tests: Set up according to Next.js testing conventions

## Architecture

### Backend (Go + Clean Architecture)

The backend follows **Clean Architecture** with strict separation of concerns:

**Layer Structure:**
1. **Domain** (`internal/domain/`): Core business entities and repository interfaces
   - `User`: Auth entity with roles (admin, editor, viewer)
   - `Tag`: Categorization entity for incidents
   - Repository interfaces defined here (e.g., `UserRepository`, `TagRepository`)

2. **Usecase** (`internal/usecase/`): Application business logic
   - `AuthUsecase`: Handles registration, login, JWT generation (24h expiry)
   - `TagUsecase`: CRUD operations for tags
   - Depends only on domain interfaces

3. **Interface** (`internal/interface/http/`): HTTP layer (Gin framework)
   - **Handlers** (`handler/`): Request/response handling
   - **Middleware** (`middleware/`): JWT authentication middleware
   - **Router** (`router/router.go`): Route registration
     - Public: `/api/auth/register`, `/api/auth/login`
     - Protected: `/api/tags/*` (requires JWT)

4. **Infrastructure** (`internal/infrastructure/persistence/`): External dependencies
   - Repository implementations using GORM
   - Database connection managed in `internal/db/db.go`

**Dependency Flow:** `main.go` → Handler → Usecase → Repository (interface) → Infrastructure (implementation)

**Key Dependencies:**
- Gin for HTTP routing
- GORM for database ORM with PostgreSQL driver
- JWT (golang-jwt/jwt/v5) for authentication
- bcrypt for password hashing

### Frontend (Next.js + TypeScript)

**Structure:**
- **App Router** (`src/app/`): Next.js 14+ pages
  - `/login`, `/signup`: Authentication pages
  - `/tags`: Tag management page
  - Root layout includes `AuthProvider`

- **Context** (`src/context/`): Global state management
  - `AuthContext`: User auth state, stored in localStorage
  - Provides `useAuth()` hook with `login()`, `logout()` methods

- **API Layer** (`src/lib/api.ts`): Centralized API client
  - `apiRequest<T>()`: Generic fetch wrapper with JWT handling
  - `authApi`: Register, login methods
  - `tagApi`: CRUD operations for tags
  - Default API URL: `http://localhost:8080/api`

- **Types** (`src/types/`): TypeScript definitions for API responses

**Tech Stack:**
- Next.js 16.0.10 with App Router
- React 19.2.1
- TailwindCSS 4 for styling
- TypeScript 5

### Infrastructure

**Services (docker-compose.yml):**
- **PostgreSQL** (port 5432): Primary database
- **Redis** (port 6379): Caching (currently defined but not actively used in code)
- **MinIO** (ports 9000/9090): S3-compatible storage (defined but not actively used)

All services connected via `incidex-network` bridge network.

## Configuration

**Backend:** Uses `internal/config/config.go` to load environment variables:
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret for JWT signing
- `PORT`: Backend server port (default 8080)

**Frontend:** Environment variables in `frontend/.env.local`:
- `NEXT_PUBLIC_API_URL`: Backend API URL (default `http://localhost:8080/api`)

## Database Migrations

Auto-migration runs on server startup in `backend/cmd/server/main.go`:
```go
dbConn.AutoMigrate(&domain.User{}, &domain.Tag{})
```

For new entities, add them to the `AutoMigrate()` call.

## Authentication Flow

1. User registers via `/api/auth/register` → User created with `viewer` role
2. User logs in via `/api/auth/login` → JWT returned (24h expiry)
3. Frontend stores JWT in localStorage via `AuthContext`
4. Protected routes require `Authorization: Bearer <token>` header
5. JWT middleware validates token and injects `userID` and `role` into Gin context

## Adding New Features

**New API Endpoint:**
1. Define entity in `internal/domain/` with repository interface
2. Create usecase in `internal/usecase/`
3. Implement repository in `internal/infrastructure/persistence/`
4. Create handler in `internal/interface/http/handler/`
5. Register route in `internal/interface/http/router/router.go`
6. Add entity to auto-migration in `cmd/server/main.go`
7. Wire dependencies in `main.go`

**Frontend Integration:**
1. Define types in `frontend/src/types/`
2. Add API methods to `frontend/src/lib/api.ts`
3. Create page/component in `frontend/src/app/` or `frontend/src/components/`

## Key Patterns

- **Error Handling (Backend):** Return errors from usecases, handlers convert to HTTP responses
- **Password Security:** Always use bcrypt, never store plaintext passwords
- **JWT Claims:** Include `user_id`, `role`, `exp` in token payload
- **API Responses:** Handlers return JSON with appropriate HTTP status codes
- **Frontend Auth:** `useAuth()` hook provides centralized auth state management
