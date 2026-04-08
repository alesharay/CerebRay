# Cerebray - Claude Code Instructions

## Project overview

Cerebray is a personal Zettelkasten web app. Users chat with AI about topics they are learning, and the AI helps format conversations into structured knowledge notes following the Zettelkasten method.

## Tech stack

- Backend: Go 1.23+ with Chi router, zerolog, sqlc, golang-migrate
- Frontend: React 19 + TypeScript + Vite + Tailwind CSS + shadcn/ui
- Database: PostgreSQL 16 with full-text search (tsvector/tsquery)
- Cache: Redis 7 (sessions, search cache)
- Auth: Keycloak OIDC with local fallback
- AI: Anthropic Claude API (streaming via SSE)
- Deployment: Docker Compose (standalone) or Kubernetes with Flux CD

## Project structure

- `backend/` - Go server
  - `cmd/server/` - entry point (main.go, router.go)
  - `db/migrations/` - golang-migrate SQL files
  - `db/queries/` - sqlc SQL queries
  - `db/sqlc/` - generated Go code
  - `internal/` - packages (ai, auth, cache, config, domain, handlers, metrics, middleware, repositories, services)
  - `docs/openapi.yaml` - API spec
- `frontend/` - React app
  - `src/api/` - HTTP client layer
  - `src/components/` - layout, ui, notes, chat, markdown
  - `src/hooks/` - custom React hooks
  - `src/pages/` - route components
  - `src/store/` - Zustand stores
  - `src/types/` - TypeScript interfaces

## Common commands

```
task infra:up          # start Postgres + Redis
task run:local         # start infra + backend + frontend
task backend:run       # Go server on :8080
task frontend:dev      # Vite dev server on :5173
task db:migrate:up     # run migrations
task db:sqlc:generate  # regenerate sqlc code
task test              # run all tests
task lint              # lint backend + frontend
```

## Conventions

- Backend handlers go in `backend/internal/handlers/`
- Database queries go in `backend/db/queries/` as sqlc SQL files
- API routes are defined in `backend/cmd/server/router.go`
- Frontend pages go in `frontend/src/pages/`
- All note content fields support Markdown
- Use structured JSON logging (zerolog) in production
- Non-root containers in Dockerfiles
