# Cerebray

> A personal Zettelkasten powered by AI conversations.

Chat with AI about topics you're learning. Cerebray helps you turn those conversations into structured, interconnected knowledge notes following the Zettelkasten method. Think of it as a second brain that actually talks back.

---

## How it works

Cerebray follows four loops that build on each other:

1. **Capture** - Chat with AI about something you're reading, watching, or thinking about. The AI asks clarifying questions and helps you articulate your understanding.
2. **Process** - Extract key ideas from conversations into atomic Zettelkasten notes. Each note gets a title, body (Markdown), and tags. The AI suggests structure and formatting.
3. **Connect** - Link related notes together. The AI suggests connections you might not see. Over time, clusters of linked notes reveal patterns in your thinking.
4. **Create** - Browse your knowledge graph, search across notes, and use your Zettelkasten as a launchpad for writing, projects, and deeper learning.

---

## Quick start (Docker Compose)

Run the full app on any machine with Docker. No Kubernetes, no private registry, no external IdP required.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (with Compose v2)
- [Task](https://taskfile.dev/installation/) (task runner)

### 1. Configure

```bash
git clone <repo-url> && cd cerebray
cp .env.example backend/.env
cp .env.example .env   # docker-compose.prod.yml reads POSTGRES_PASSWORD from here
```

Edit `backend/.env` and set these values at minimum:

| Variable | What to set | Example |
|----------|------------|---------|
| `SESSION_SECRET` | Any random string, 32+ characters | `openssl rand -hex 32` |
| `ANTHROPIC_API_KEY` | API key from [console.anthropic.com](https://console.anthropic.com) | `sk-ant-...` |
| `POSTGRES_PASSWORD` | Database password, in the repo-root `.env` | any strong string |

`DATABASE_URL` and `REDIS_URL` are set by `docker-compose.prod.yml` itself, so
editing them in `backend/.env` has no effect on the Compose stack. They matter
only when running the backend directly with `task backend:run`.

Leave `KEYCLOAK_ISSUER_URL` empty to use local auth (no Keycloak needed). See the [Authentication](#authentication) section for details.

### 2. Start

```bash
task prod:up
```

The app is running at `http://localhost`. Click "Sign in" on the landing page to create a local admin user and start taking notes.

```bash
task prod:logs     # tail all logs
task prod:down     # stop everything
task prod:reset    # stop and wipe all data
```

---

## Authentication

The app supports two auth modes, selected automatically based on your configuration.

**Local mode** (default) - When `KEYCLOAK_ISSUER_URL` is empty, the app runs in local auth mode. A single admin user is created on first login. No external identity provider needed. Set `LOCAL_USER_EMAIL` and `LOCAL_USER_NAME` in your `.env` to customize the user (defaults to `admin@localhost` / `Local Admin`).

**Keycloak OIDC mode** - Set `KEYCLOAK_ISSUER_URL`, `KEYCLOAK_CLIENT_ID`, and `KEYCLOAK_CLIENT_SECRET` in your `.env` to enable Keycloak SSO. Users are created in PostgreSQL on first OIDC login and updated on subsequent logins.

The session mechanism is identical in both modes. Downstream handlers don't know or care which auth flow created the session.

---

## Tech stack

| Layer | Choice |
| --- | --- |
| Frontend | Vite + React 19 + TypeScript + Tailwind CSS (shadcn utilities only) |
| State | Zustand |
| Backend | Go with Chi router |
| Auth | Keycloak OIDC or local auth, sessions in Redis |
| AI | Anthropic Claude (streaming via SSE) |
| Database | PostgreSQL, full-text search with tsvector/tsquery (16 locally, 18 in the cluster) |
| Cache | Redis, sessions only (7 locally, 8 in the cluster) |
| SQL | sqlc (type-safe generated Go from SQL) |
| Migrations | golang-migrate |
| CI | Gitea Actions (lint, test, build, push on push to main) |
| Observability | Prometheus `/metrics`, Grafana Faro RUM |

---

## Local development

### Prerequisites

- Go 1.26+
- Node 22+
- Docker (for PostgreSQL and Redis)
- [Task](https://taskfile.dev/installation/) (task runner)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

sqlc does not need installing - `task db:sqlc:generate` runs a pinned version
via `go run`, so everyone generates identical code. See `SQLC_VERSION` in
`Taskfile.yml` before bumping it.

### First-time setup

```bash
cp .env.example backend/.env
# Fill in at minimum: ANTHROPIC_API_KEY, SESSION_SECRET

task frontend:install      # install npm dependencies
task backend:tidy          # tidy Go modules
task infra:up              # start PostgreSQL and Redis
task db:migrate:up         # run database migrations
```

### Daily dev workflow

```bash
task infra:up              # start PostgreSQL and Redis (if not already running)
task backend:run           # start Go server on :8080 (separate terminal)
task frontend:dev          # start Vite dev server on :5173 (separate terminal)
```

Or run everything at once:

```bash
task run:local             # starts infra + backend + frontend concurrently
```

The Vite dev server proxies `/api` and `/auth` requests to `:8080` automatically. Visit `http://localhost:5173`.

### Run tests

```bash
task test                  # run all backend and frontend tests in parallel
task backend:test          # Go tests only
task backend:test:short    # skip long-running integration tests
task backend:test:integration  # testcontainers integration tests (needs Docker)
task frontend:test         # Vitest component tests only
task frontend:test:e2e     # Playwright end-to-end tests
```

### Other useful commands

```bash
task lint                  # lint backend (go vet) and frontend (eslint)
task typecheck             # go vet + tsc --noEmit
task build                 # compile Go binary + Vite production build
task db:migrate:create -- add_tags_table   # create a new migration
task db:sqlc:generate      # regenerate Go code from SQL queries
task db:seed               # load sample data for local dev
task infra:logs            # tail Docker Compose logs
task infra:reset           # wipe local DB volumes (destructive)
task                       # list all available tasks
```

### Cluster operations

```bash
task k8s:diagnose          # nodes + pods + events + storage, start here when something is down
task k8s:nodes             # node status
task k8s:events            # recent namespace events, oldest first
task k8s:storage           # PVCs, PVs and storage classes
task k8s:logs:backend      # tail backend logs
task k8s:logs:backend:prev # previous container's logs (use when crash looping)
task k8s:manifests         # regenerate gitops manifests (guarded against a stale checkout)
task flux:status           # Flux reconciliation state
```

---

## Project structure

```text
cerebray/
├── backend/
│   ├── cmd/server/          # Entry point (main.go, router.go)
│   ├── db/
│   │   ├── migrations/      # SQL migration files
│   │   ├── queries/         # sqlc query files
│   │   └── sqlc/            # Generated Go code
│   ├── Dockerfile
│   └── internal/
│       ├── ai/              # Claude API client + streaming
│       ├── auth/            # Keycloak OIDC, local auth, sessions in Redis
│       ├── config/          # Environment config loading
│       ├── handlers/        # HTTP handlers (use db/sqlc directly)
│       ├── metrics/         # Prometheus collectors
│       └── middleware/      # Logging, recovery, CORS, auth
├── frontend/
│   ├── src/
│   │   ├── api/             # HTTP client layer
│   │   ├── components/      # UI components (layout, notes)
│   │   ├── hooks/           # Custom React hooks
│   │   ├── lib/             # utils, zettelParser
│   │   ├── pages/           # Route-level page components
│   │   ├── store/           # Zustand stores
│   │   ├── types/           # TypeScript type definitions
│   │   └── faro.ts          # Grafana Faro RUM setup
│   ├── e2e/                 # Playwright end-to-end tests
│   ├── nginx.conf           # Production reverse proxy config
│   ├── Dockerfile
│   └── package.json
├── .gitea/workflows/ci.yaml # CI pipeline
├── .env.example             # App secrets template
├── docker-compose.dev.yml   # Local dev infrastructure (PostgreSQL + Redis)
├── docker-compose.prod.yml  # Full production stack
├── deploy.env.example       # Deployment config overrides
├── Taskfile.yml             # Task runner (dev, test, build, deploy)
├── troubleshooting.md       # Issue resolution log
└── PROJECT_PLAN.md          # Build plan and progress tracker
```

There is no repository layer and no service layer - handlers call the sqlc
generated queries directly. There is no OpenAPI spec yet; `backend/cmd/server/router.go`
is the authoritative list of routes.

---

## Deployment

### Docker Compose (simplest)

See the [Quick start](#quick-start-docker-compose) section above. Clone, configure `.env`, run `task prod:up`. Works on any machine with Docker.

### Kubernetes (GitOps)

The deployment targets a k3d cluster with Flux CD. All deployment-specific values (registry, domain, cert issuer, Vault path) are parameterized in the Taskfile with homelab defaults. To deploy to your own k8s cluster:

```bash
cp deploy.env.example deploy.env
# Set REGISTRY, APP_DOMAIN, CERT_ISSUER, etc. for your environment
```

Key overrides:

| Variable | What it controls | Example |
|----------|-----------------|---------|
| `REGISTRY` | Docker image registry | `ghcr.io/your-org` |
| `APP_DOMAIN` | Public domain for ingress | `cerebray.example.com` |
| `CERT_ISSUER` | cert-manager ClusterIssuer for TLS | `letsencrypt-prod` |
| `VAULT_SECRET_STORE` | ExternalSecrets ClusterSecretStore name | `your-vault-store` |
| `VAULT_PATH` | Vault KV path for secrets | `secret/cerebray` |
| `GITOPS_DIR` | Path to your gitops repo | `../your-gitops-repo` |

The Taskfile loads `deploy.env` automatically. After setting overrides, `task k8s:manifests` generates k8s YAML and `task deploy:all` builds and pushes images.

See [PROJECT_PLAN.md](PROJECT_PLAN.md) Phase 13 for the full deployment checklist and [troubleshooting.md](troubleshooting.md) for resolved issues.

---

## Status

See [PROJECT_PLAN.md](PROJECT_PLAN.md) for current build progress.
