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
- [x] golang-migrate wired up, all 8 migrations created and applied
- [x] sqlc configured, queries for all tables (users, notes, tags, connections, conversations, glossary, ai_usage)
- [x] Schema defined: users, notes, tags/note_tags, connections, conversations/messages, glossary_terms, ai_usage_log
- [x] Full-text search via tsvector/tsquery on note title and body
- [x] Seed script for local development (task db:seed)

### 4. Auth (Keycloak OIDC + Local Fallback)

- [x] Keycloak OIDC integration (oauth2 + PKCE)
- [x] Session management via Redis (custom store with sliding expiry)
- [x] Local auth fallback when KEYCLOAK_ISSUER_URL is empty
- [x] Auth middleware protecting /api routes
- [x] Login and logout flow tested end to end (Go integration test with miniredis)

### 5. Backend Core

- [x] Chi router with middleware (logging, recovery, CORS, auth)
- [x] Health check endpoint (/health)
- [x] Zerolog structured logging
- [x] Config loaded from environment (godotenv + fail-fast validation)
- [x] CRUD handlers for notes
- [x] CRUD handlers for tags
- [x] Note linking (bidirectional references)
- [x] Search endpoint using PostgreSQL full-text search

### 6. AI Integration

- [x] Anthropic Claude API client with streaming (SSE)
- [x] Chat endpoint (POST /conversations/:id/messages) that streams responses
- [x] System prompt for Zettelkasten formatting with ZETTEL_SUGGESTION blocks
- [x] Token budget tracking per user per month (GET /settings/usage)
- [x] Conversation history stored in database
- [x] "Extract note" action from chat messages (text selection with floating button)

### 7. Frontend Foundation

- [x] Vite + React 19 + TypeScript scaffold
- [x] Tailwind CSS + shadcn/ui utilities (cn, lucide-react)
- [x] React Router with protected routes
- [x] Zustand stores (auth)
- [x] HTTP client layer with auth redirect on 401
- [x] Layout components (Sidebar, AppLayout with Outlet)

### 8. Frontend Pages

- [x] Landing page (intro + Keycloak login button)
- [x] Dashboard (stat cards for Inbox/Echoes/Codex, quick capture, recent activity)
- [x] Chat / Capture page (AI conversation with Zettel suggestion cards)
- [x] Inbox page (fleeting notes with promote/sleep/archive actions)
- [x] Echoes page (sleeping notes with Rise/Rest actions)
- [x] Codex page (main library with search, filters, card/list view)
- [x] Note Detail / Editor page (all Codex fields, connection builder, tags)
- [x] Index / MOC page (list view + d3-force graph with toggle)
- [x] Glossary page (alphabetical terms with definitions)
- [x] Settings page (profile, AI usage, display preferences)

### 9. Search and Connections

- [x] Full-text search with highlighted results (ts_headline snippets)
- [x] Tag filtering and combination (tag chips on Codex page)
- [x] Backlink display on note pages (outgoing/incoming split)
- [x] Forward link suggestions from AI (CONNECTION_SUGGESTION blocks)
- [x] Graph visualization of note connections (d3-force interactive graph)

### 10. Analytics Dashboard (Front Page)

The dashboard is the main landing page after login. It surfaces patterns in your
Zettelkasten usage so you can see how your knowledge base is growing without
digging through individual notes. All lifecycle data (timestamps, status
transitions, dwell times) is tracked automatically by the backend - nothing here
requires manual entry.

#### 10a. Schema - note_events table

- [x] New migration: `note_events` table (note_id, user_id, from_status, to_status, created_at)
- [x] Backend logs an event on every note status transition (promote, sleep, archive, rise)
- [x] Backfill initial "created" events from existing notes.created_at

#### 10b. Backend analytics endpoint

- [x] GET /api/v1/dashboard/analytics returns aggregated metrics
- [x] Inbox snapshot: fleeting notes with age and readiness score (fields filled / total fields)
- [x] Lifecycle metrics: avg dwell time in inbox, promote/sleep/archive ratios, trend over time
- [x] Zettelkasten strength score composed of:
  - Connection density (avg connections per active/linked note)
  - Orphan count (active notes with zero connections)
  - Glossary coverage (glossary terms / total active notes)
  - Type diversity (distribution across note_type enum values)
  - Sleeping note backlog (count + avg age)
- [x] Conversation-to-note conversion rate (conversations with at least one source_chat_id note / total)
- [x] AI usage budget widget data (monthly spend vs budget from ai_usage_log)
- [x] Stale note detection (active notes not updated in N days)

#### 10c. Frontend dashboard redesign

- [x] Replace current stat-card dashboard with analytics-driven front page
- [x] Inbox section: list of fleeting notes with age badges + readiness indicators
- [x] Lifecycle section: promote/sleep/archive ratios, avg triage time, trend bar
- [x] Zettelkasten health section: strength ring score with metric breakdown
- [x] Conversation conversion rate widget
- [x] AI usage widget
- [x] Stale note nudges (notes that need attention)
- [x] Trend sparklines for lifecycle data over time

#### 10d. UI identity

Cerebray has its own visual identity, distinct from Grafana-style monitoring UIs.
Design principles for the dashboard and all pages:

- [x] Quicksand as the primary font family
- Warm, approachable palette - not cold infrastructure grays
- Typographic hierarchy over dense data grids
- Generous whitespace; the UI should breathe
- Cards and sections feel like a notebook or study companion, not a dashboard panel
- Subtle animations and transitions where they add clarity
- Data visualizations should be clean and minimal, not chart-heavy

### 10e. Interactive Knowledge Graph (Index Page)

- [x] Zoom and pan (d3-zoom) on the force-directed graph
- [x] Click a node to navigate to the note detail page
- [x] Hover tooltip showing note title, type, and connection count
- [x] Drag nodes to rearrange the layout
- [x] Color legend for note types
- [x] Highlight connected nodes on hover (dim unrelated nodes)
- [x] Search/filter within the graph view
- [ ] Cluster visualization (group tightly connected notes) - deferred, revisit at 100+ notes

### 10f. UX Fixes and Polish

- [x] Dashboard: fleeting notes link to inbox, not codex
- [x] Echoes page: simplify layout to title + age + actions (promote/archive), not full Zettel fields
- [x] Chat suggestions: show "Save to Inbox" cards when AI suggests new notes during follow-up chat
- [x] Save button rework: "Refresh from chat" action that re-processes the conversation and updates the note
- [x] Remove debug console.log statements from NoteDetailPage
- [x] Trend sparklines for lifecycle data over time (dashboard)

### 11. Testing

- [x] Unit tests for Go handlers (table-driven with fake Querier)
- [x] Unit tests for auth middleware
- [x] Integration tests against live PostgreSQL (testcontainers, build-tagged)
- [x] Frontend testing infrastructure (Vitest + RTL + jsdom configured)
- [x] Frontend component tests (Vitest + RTL) - parser, NoteCard, Sidebar, Dashboard, Echoes, Landing, authStore, utils
- [x] End-to-end tests (Playwright smoke specs)
- [x] Smoke test against live deployment (task test:smoke)

### 12. Observability

- [x] Prometheus /metrics endpoint (HTTP request counters, histograms)
- [x] AI usage metrics (tokens consumed, requests, latency)
- [x] Frontend RUM via Grafana Faro (`src/faro.ts`). PostHog dropped.
- [x] Structured JSON logging in production (zerolog, request_id correlation)
- [x] Scrape config for cerebray. Pod annotations were added then removed - this cluster discovers via ServiceMonitors. A ServiceMonitor now exists and Prometheus reports the target UP.

### 13. Deployment

- [x] Backend Dockerfile (multi-stage Go build, Alpine runtime)
- [x] Frontend Dockerfile (Vite build + nginx:alpine runtime)
- [x] docker-compose.prod.yml (full production stack)
- [x] Images pushed to registry.homelab
- [x] Flux CD image automation (ts-* tag pattern, ImageRepositories + ImagePolicies)
- [x] k8s manifests written to homelab-gitops/apps/base/cerebray/
- [x] PostgreSQL HelmRelease (Bitnami chart)
- [x] Redis HelmRelease (Bitnami chart)
- [x] Vault secrets pushed via task secrets:vault
- [x] ExternalSecrets pulling from Vault
- [x] Ingress with TLS via cert-manager
- [x] Database migrations applied to production PostgreSQL
- [x] All pods running and healthy

---

## Tech Stack Summary

| Layer | Choice | Notes |
| --- | --- | --- |
| Frontend framework | Vite + React 19 + TypeScript | |
| Styling | Tailwind CSS + shadcn utilities (cn, lucide-react) | No component library |
| State | Zustand | |
| Routing | React Router v7 | |
| HTTP client | fetch / custom wrapper | |
| Frontend tests | Vitest + RTL + Playwright | |
| Backend language | Go 1.26+ | |
| Backend router | Chi | Lightweight, idiomatic, middleware-friendly |
| Auth | Keycloak OIDC via coreos/go-oidc | Session in Redis, local fallback mode |
| AI engine | Anthropic Claude API | SSE streaming, Zettelkasten system prompt |
| Database | PostgreSQL via sqlc + golang-migrate | Full-text search. 16 locally, 18 in the cluster |
| Cache | Redis | Sessions only. 7 locally, 8 in the cluster |
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
│   │   ├── config/
│   │   ├── handlers/
│   │   ├── metrics/
│   │   └── middleware/
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   └── notes/
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── pages/
│   │   ├── store/
│   │   ├── types/
│   │   └── faro.ts
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

## Deferred / next up

- [ ] **OpenAPI spec + Swagger UI.** There is no spec today; `backend/cmd/server/router.go`
      is the source of truth for all 29 routes. The goal is a browsable Swagger page so the
      API can be exercised and documented for users. Sized as its own piece of work.
- [x] **Database backups.** Nightly CronJobs for all four databases (cerebray, archdraft
      and keycloak Postgres, do-a-doc MongoDB) dump to the NAS over NFS with 14 day
      retention, staggered an hour apart. Verified by restoring into a scratch database,
      not just by the job exiting 0. `task k8s:backup:now|list|verify` drive them by hand.
- [x] **Fix the collation version mismatch.** `REINDEX DATABASE cerebray` plus
      `ALTER DATABASE cerebray REFRESH COLLATION VERSION` rebuilt all 31 indexes under
      glibc 2.43. Keycloak and archdraft were surveyed and are unaffected. The `postgres`
      and `template1` databases remain stale because this deployment has no superuser
      password; the only cost is that `CREATE DATABASE` needs `TEMPLATE template0`.
- [x] **Pin container image tags.** Every workload in the cluster is now pinned; a sweep
      for `:latest` returns nothing. bitnami's postgresql, redis and mongodb no longer
      publish versioned tags at all, so those are pinned by digest to the builds already
      running. bitnamilegacy (keycloak) still has real tags, verified byte-identical to
      the `latest` they replaced. Chart versions are pinned exactly too - a floating range
      lets the chart drift underneath a pinned image. The cerebray postgresql chart stays
      on 16.7.27 despite driving a PG 18.6 image; moving to an 18.x chart is its own job.
- [ ] **Move stateful workloads to the `nfs` StorageClass.** `local-path` pins a PV to a
      single node, which is what caused the 2026-09-20 outage. Keycloak already uses `nfs`.
      Currently recommended against for cerebray: the data is tiny, the PV is now `Retain`,
      backups exist, and that StorageClass is `nfsvers=3,nolock` - weak fsync durability
      and no locking under a single-writer database.
      Requires a dump and restore since a bound PVC cannot be retargeted.
- [ ] **Markdown rendering.** Note fields store Markdown but the UI renders raw text.
- [x] **ServiceMonitor for cerebray.** Prometheus scrapes the backend every 30s. Needed
      two labels: `release: kube-prometheus-stack` on the ServiceMonitor, which is what the
      Prometheus CR selects on, and `app: backend` on the *Service metadata* - a
      ServiceMonitor selector matches Service labels, not `spec.selector`.
- [ ] **Migration Job in the deploy path.** Migrations are applied by hand today.
- [x] Raise `fs.inotify.max_user_instances` persistently. Two layers: a DaemonSet in
      `homelab-gitops/infrastructure/custom-resources/sysctl-inotify.yaml` that Flux
      reapplies every reconcile (512 -> 8192 on all three nodes, verified), and a
      `provision:` block in `~/.colima/default/colima.yaml` so a VM restart keeps it.

## Notes and Decisions Log

- **2026-09-20**: Full drift audit after the Postgres outage (see troubleshooting.md).
  Fixed: tag creation had never worked (frontend sent `{name}`, backend decoded `{tags:[]}`,
  0 tag rows in production) and now has regression tests; search snippets were shipping as
  base64 because sqlc inferred `[]byte` from `ts_headline`; connection and note-tag queries
  were not scoped by `user_id`; sqlc is pinned to v1.30.0 via `go run` since v1.31 changes
  nullable enum types and breaks the build; `schema_migrations` did not exist in production
  and was baselined to 8; all six PVs patched to `reclaimPolicy: Retain`; the `k8s:manifests`
  template had five regressions (SSL_CERT_FILE, KEYCLOAK_ISSUER_URL, KEYCLOAK_CLIENT_ID, the
  homelab-ca mount, and a 256Mi memory limit that would OOMKill SSE streams) plus a guard so
  it refuses to write into a stale gitops checkout. Removed dead code: `internal/domain` was
  imported by nothing and was the source of the TypeScript/Go type divergence; PostHog build
  args were plumbed through four files and consumed by none. Vault was found sealed, which
  had silently broken every ExternalSecret in the cluster - all seven now sync.

- **2026-04-10**: Phase 11 (Testing) mostly complete. Enabled sqlc Querier interface for handler mocking. Backend: 28 handler unit tests (notes CRUD/promote/search, chat usage, health, helpers) + 5 auth middleware tests with miniredis. Frontend: 43 tests across 8 files (zettel parser, NoteCard, Sidebar, DashboardPage, EchoesPage, LandingPage, authStore, utils). Added smoke test Taskfile task. Integration tests (testcontainers) and Playwright e2e have since landed - see `backend/internal/handlers/integration_test.go` and `frontend/e2e/`.
- **2026-04-10**: Phase 12 (Observability) complete. Added Prometheus metrics via `internal/metrics/` package: HTTP request counter/histogram/gauge and AI token/request/duration metrics. Chi middleware records HTTP metrics using route patterns for low cardinality. `/metrics` endpoint exposed unauthenticated for Prometheus scraping. Enhanced structured logging with `request_id` correlation in all request logs, plus `RequestLogger` context helper for handler-level logs with user_id. PostHog skipped. Correction (2026-09-20): the pod annotations were never applied and would be inert anyway - this cluster discovers targets via ServiceMonitors, and no ServiceMonitor existed for cerebray. Resolved 2026-09-20 - one now exists and the target reports UP. `/metrics` is still not routed by the ingress, so it is reachable in-cluster only.
- **2026-04-10**: Phase 10e/10f complete. Interactive knowledge graph with zoom/pan (d3-zoom), drag (d3-drag), hover highlighting, HTML tooltips, color legend, search/filter, and SPA navigation. Cluster visualization deferred until 100+ notes. Echoes page simplified to title + age + actions. Chat follow-up suggestions with Save to Inbox cards. Refresh from chat action on note detail page. Dashboard trend sparklines via new GetLifecycleTrend backend query (weekly counts over 90 days).
- **2026-04-10**: Workflow refactor complete (Phase 10). Inbox is now quick-capture, promote triggers AI expansion, note detail page has embedded chat. Chat page removed from nav. Added Phase 10e (interactive knowledge graph) and 10f (UX fixes) based on user testing. Broken d3 zoom placeholder fixed. SSE buffer flush bug found and fixed (done event not processed). Zettel parser upgraded for multi-line field content. Deployed to homelab k8s with all infrastructure operational.
- **2026-04-09**: Added Phase 10 (Analytics Dashboard). New `note_events` table to track status transitions automatically. Dashboard redesign with inbox overview, lifecycle metrics, Zettelkasten strength score, conversation conversion rate, AI budget, and stale note detection. All lifecycle data is system-tracked, no manual entry. UI identity established: Quicksand font, warm palette, notebook aesthetic - explicitly not a Grafana-style monitoring UI.
- **2026-04-08**: Project plan created. Tech stack confirmed. Chi chosen over Gin for idiomatic Go alignment. PostgreSQL chosen over MongoDB for structured Zettelkasten data with full-text search (tsvector/tsquery). Keycloak OIDC chosen over Okta SAML for open-source self-hostable auth. sqlc chosen for type-safe SQL over an ORM. Claude API with SSE streaming for AI-assisted note creation.
