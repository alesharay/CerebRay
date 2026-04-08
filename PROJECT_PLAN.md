# Cerebray - Project Plan

## Status Legend

- [ ] Not started
- [~] In progress
- [x] Complete

## Phases

### 1. Project Setup

- [x] Create project directory structure
- [x] CLAUDE.md with project conventions
- [x] .gitignore, .env.example, deploy.env.example
- [x] Taskfile.yml with all task namespaces
- [x] docker-compose.dev.yml (PostgreSQL + Redis)
- [x] PROJECT_PLAN.md and README.md
- [x] Initialize Go module and frontend scaffold
- [x] troubleshooting.md

### 2. Git Setup

- [x] Initialize git repo
- [x] Initial commit with project scaffold
- [x] Set up Gitea and GitHub remotes
- [x] CI pipeline (Gitea Actions: lint, test, build on push)

### 3. Database

- [x] PostgreSQL running via docker-compose.dev.yml
- [x] golang-migrate wired up, all 7 migrations created and applied
- [x] sqlc configured, queries for all tables (users, notes, tags, connections, conversations, glossary, ai_usage)
- [x] Schema defined: users, notes, tags/note_tags, connections, conversations/messages, glossary_terms, ai_usage_log
- [x] Full-text search via tsvector/tsquery on note title and body
- [ ] Seed script for local development

### 4. Auth (Keycloak OIDC + Local Fallback)

- [ ] Keycloak OIDC integration (coreos/go-oidc + oauth2)
- [ ] Session management via Redis (gorilla/sessions or similar)
- [ ] Local auth fallback when KEYCLOAK_ISSUER_URL is empty
- [ ] Auth middleware protecting /api routes
- [ ] Login and logout flow tested end to end

### 5. Backend Core

- [ ] Chi router with middleware (logging, recovery, CORS, auth)
- [ ] Health check endpoint (/health)
- [ ] Zerolog structured logging
- [ ] Config loaded from environment (envconfig or similar)
- [ ] CRUD handlers for notes
- [ ] CRUD handlers for tags
- [ ] Note linking (bidirectional references)
- [ ] Search endpoint using PostgreSQL full-text search

### 6. AI Integration

- [ ] Anthropic Claude API client with streaming (SSE)
- [ ] Chat endpoint (/api/chat) that streams responses
- [ ] System prompt for Zettelkasten formatting
- [ ] Token budget tracking per user per month
- [ ] Conversation history stored in database
- [ ] "Extract note" action from chat messages

### 7. Frontend Foundation

- [ ] Vite + React 19 + TypeScript scaffold
- [ ] Tailwind CSS + shadcn/ui component library
- [ ] React Router with protected routes
- [ ] Zustand stores (auth, notes, chat)
- [ ] HTTP client layer with auth interceptor
- [ ] Layout components (sidebar, header, main content area)

### 8. Frontend Pages

- [ ] Landing page (intro + Keycloak login button)
- [ ] Dashboard (stat cards for Inbox/Echoes/Codex, quick capture, recent activity)
- [ ] Chat / Capture page (AI conversation with Zettel suggestion cards)
- [ ] Inbox page (fleeting notes with promote/sleep/archive actions)
- [ ] Echoes page (sleeping notes with Rise/Rest actions)
- [ ] Codex page (main library with search, filters, card/list view)
- [ ] Note Detail / Editor page (all Codex fields, connection builder, tags)
- [ ] Index / MOC page (hierarchical list + react-force-graph visualization)
- [ ] Glossary page (alphabetical terms with definitions)
- [ ] Settings page (profile, AI usage, display preferences)

### 9. Search and Connections

- [ ] Full-text search with highlighted results
- [ ] Tag filtering and combination
- [ ] Backlink display on note pages
- [ ] Forward link suggestions from AI
- [ ] Graph visualization of note connections (D3 or similar)

### 10. Testing

- [ ] Unit tests for Go handlers (table-driven with fake stores)
- [ ] Unit tests for auth middleware
- [ ] Integration tests against live PostgreSQL
- [ ] Frontend component tests (Vitest + RTL)
- [ ] End-to-end tests (Playwright)
- [ ] Smoke test against live deployment

### 11. Observability

- [ ] Prometheus /metrics endpoint (HTTP request counters, histograms)
- [ ] AI usage metrics (tokens consumed, requests, latency)
- [ ] PostHog analytics (page views, note creation, chat usage)
- [ ] Structured JSON logging in production (zerolog)

### 12. Deployment

- [ ] Backend Dockerfile (multi-stage Go build, Alpine runtime)
- [ ] Frontend Dockerfile (Vite build + nginx:alpine runtime)
- [ ] docker-compose.prod.yml (full production stack)
- [ ] Images pushed to registry.homelab
- [ ] Flux CD image automation (ts-* tag pattern)
- [ ] k8s manifests generated via task k8s:manifests
- [ ] PostgreSQL HelmRelease (Bitnami chart)
- [ ] Redis HelmRelease (Bitnami chart)
- [ ] Vault secrets pushed via task secrets:vault
- [ ] ExternalSecrets pulling from Vault
- [ ] Ingress with TLS via cert-manager
- [ ] All pods running and healthy

---

## Tech Stack Summary

| Layer | Choice | Notes |
| --- | --- | --- |
| Frontend framework | Vite + React 19 + TypeScript | |
| Styling | Tailwind CSS + shadcn/ui | |
| State | Zustand | |
| Routing | React Router v7 | |
| HTTP client | fetch / custom wrapper | |
| Frontend tests | Vitest + RTL + Playwright | |
| Backend language | Go 1.23+ | |
| Backend router | Chi | Lightweight, idiomatic, middleware-friendly |
| Auth | Keycloak OIDC via coreos/go-oidc | Session in Redis, local fallback mode |
| AI engine | Anthropic Claude API | SSE streaming, Zettelkasten system prompt |
| Database | PostgreSQL 16 via sqlc + golang-migrate | Full-text search with tsvector/tsquery |
| Cache | Redis 7 | Sessions, search cache |
| Deployment | Docker Compose or k8s with Flux CD | GitOps two-repo pattern |
| Secrets | Vault + ExternalSecrets operator | Vault on NAS, ESO syncs to k8s |
| Ingress | ingress-nginx | Routes /api, /auth to backend; / to frontend |
| TLS | cert-manager | homelab-ca-issuer for internal, letsencrypt for prod |
| Image registry | registry.homelab (NAS) | Flux Image Automation watches for ts-* tags |
| Local dev | Docker Compose (docker-compose.dev.yml) | PostgreSQL + Redis for host-based dev |

### Router Choice: Chi over Gin

Chi is recommended for this project. It is stdlib-compatible (uses net/http natively), has zero external dependencies beyond the router itself, and composes cleanly with standard middleware. Gin is excellent but has a custom Context type that diverges from stdlib and adds conceptual overhead for someone learning Go. Chi lets you write idiomatic handlers that look like standard library code.

---

## Directory Structure (target)

```text
cerebray/
├── PROJECT_PLAN.md
├── README.md
├── CLAUDE.md
├── .gitignore
├── .env.example
├── deploy.env.example
├── docker-compose.dev.yml
├── docker-compose.prod.yml
├── Taskfile.yml
├── troubleshooting.md
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       ├── main.go
│   │       └── router.go
│   ├── db/
│   │   ├── migrations/
│   │   ├── queries/
│   │   └── sqlc/
│   ├── internal/
│   │   ├── ai/
│   │   ├── auth/
│   │   ├── cache/
│   │   ├── config/
│   │   ├── domain/
│   │   ├── handlers/
│   │   ├── metrics/
│   │   ├── middleware/
│   │   ├── repositories/
│   │   └── services/
│   ├── docs/
│   │   └── openapi.yaml
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   ├── ui/
│   │   │   ├── notes/
│   │   │   ├── chat/
│   │   │   └── markdown/
│   │   ├── hooks/
│   │   ├── pages/
│   │   ├── store/
│   │   └── types/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── nginx.conf
│   └── package.json
```

K8s manifests live in a separate repo (`homelab-gitops/apps/base/cerebray/`):
backend.yaml, frontend.yaml, postgresql.yaml, redis.yaml,
ingress.yaml, secret.yaml, namespace.yaml, kustomization.yaml

---

## Notes and Decisions Log

- **2026-04-08**: Project plan created. Tech stack confirmed. Chi chosen over Gin for idiomatic Go alignment. PostgreSQL chosen over MongoDB for structured Zettelkasten data with full-text search (tsvector/tsquery). Keycloak OIDC chosen over Okta SAML for open-source self-hostable auth. sqlc chosen for type-safe SQL over an ORM. Claude API with SSE streaming for AI-assisted note creation.
