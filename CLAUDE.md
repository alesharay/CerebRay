# Cerebray - Claude Code Instructions

## Project overview

Cerebray is a personal Zettelkasten web app. Users chat with AI about topics they are learning, and the AI helps format conversations into structured knowledge notes following the Zettelkasten method.

## Tech stack

- Backend: Go 1.26+ with Chi router, zerolog, sqlc, golang-migrate
- Frontend: React 19 + TypeScript + Vite + Tailwind CSS (shadcn utilities only - `cn()` and lucide-react, no component library)
- Database: PostgreSQL with full-text search (tsvector/tsquery). 16 in local Compose, 18 in the homelab cluster - see troubleshooting.md before changing either pin.
- Cache: Redis (sessions only). Local Compose runs 7; the cluster runs 8 via an unpinned chart tag.
- Auth: Keycloak OIDC with local fallback
- AI: Anthropic Claude API (streaming via SSE)
- Observability: Prometheus metrics at `/metrics`, Grafana Faro RUM in the frontend
- Deployment: Docker Compose (standalone) or Kubernetes with Flux CD

## Project structure

- `backend/` - Go server
  - `cmd/server/` - entry point (main.go, router.go)
  - `db/migrations/` - golang-migrate SQL files
  - `db/queries/` - sqlc SQL queries
  - `db/sqlc/` - generated Go code, used directly by handlers (no repository layer)
  - `internal/` - packages (ai, auth, config, handlers, metrics, middleware)
- `frontend/` - React app
  - `src/api/` - HTTP client layer
  - `src/components/` - layout, notes
  - `src/hooks/` - custom React hooks
  - `src/lib/` - utils, zettelParser
  - `src/pages/` - route components
  - `src/store/` - Zustand stores
  - `src/types/` - TypeScript interfaces

There is no OpenAPI spec yet. Routes are defined in `backend/cmd/server/router.go`,
which is the source of truth until a spec exists.

## Common commands

```
task infra:up          # start Postgres + Redis
task run:local         # start infra + backend + frontend
task backend:run       # Go server on :8080
task frontend:dev      # Vite dev server on :5173
task db:migrate:up     # run migrations
task db:sqlc:generate  # regenerate sqlc code (pinned to SQLC_VERSION)
task db:seed           # seed sample data for local dev
task test              # run all tests
task lint              # lint backend + frontend

task k8s:diagnose      # first command when something is down in the cluster
task k8s:logs:backend:prev  # backend logs from the previous container (crash loops)
```

## Conventions

- Backend handlers go in `backend/internal/handlers/`
- Database queries go in `backend/db/queries/` as sqlc SQL files
- API routes are defined in `backend/cmd/server/router.go`
- Frontend pages go in `frontend/src/pages/`
- Note content fields are stored as Markdown. Nothing renders it yet - the UI
  displays the raw text, so a Markdown renderer is still outstanding.
- Queries that read or write user-owned rows must scope by `user_id` in SQL,
  not just in the handler
- Use structured JSON logging (zerolog) in production
- Non-root containers in Dockerfiles
