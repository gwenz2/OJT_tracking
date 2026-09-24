# OJT Attendance & Daily Journal Management System

Mobile-first attendance and daily journal tracking for on-the-job training (OJT)
programs. Trainees time in/out with live camera evidence and GPS; coordinators
monitor progress, review journals, and approve corrections; reports export to CSV.

**Architecture**: Go (Fiber v3) modular monolith + PostgreSQL + S3-compatible
private object storage (MinIO in dev) + React 19 / TypeScript / Vite PWA.
Opaque cookie sessions with CSRF protection — no JWT, no offline attendance
authority. Full design: `specs/001-ojt-attendance-daily-journal/`.

## Repository layout

```
backend/            Go API (Fiber v3, pgx v5)
  cmd/api           server + --migrate-up/--migrate-down/--seed-admin/--reminders
  cmd/seed          dev demo data (trainee/site/assignment)
  internal/         app, auth, trainees, sites, assignments, attendance,
                    journals, corrections, notifications, reports, settings,
                    users, audit, platform/{db,httpx,storage,observability,ratelimit}
  tests/            contract + integration tests (real Postgres/MinIO)
frontend/           React 19 + Vite PWA (TanStack Query, RHF+Zod, Tailwind v4)
specs/              feature spec, plan, data model, API contracts, tasks
archive/            retired generic template (not part of the product)
docker-compose.yml  local PostgreSQL 18 + MinIO
```

## Local development

Prereqs: Go 1.26+, Node 24+, Docker (for Postgres + MinIO), `make` optional.

```bash
# 1. dependencies
docker compose up -d                       # postgres :5432, minio :9000/:9001

# 2. backend env + run (migrations apply at startup)
cp backend/.env.example backend/.env       # defaults match docker-compose
cd backend && go run ./cmd/api --seed-admin   # ADMIN_EMAIL/ADMIN_PASSWORD env
go run ./cmd/api                            # http://localhost:8080

# 3. frontend
cd frontend && npm ci && npm run dev        # http://localhost:5173
```

`make` shortcuts exist for every step (`make compose-up backend frontend test`).

## Environments

| Env | API | DB | Storage | Notes |
|---|---|---|---|---|
| dev | `localhost:8080` | docker compose postgres | MinIO `ojt-evidence` bucket, auto-created | cookies are non-Secure (HTTP) |
| staging | same-origin behind TLS | managed postgres | S3-compatible private bucket | Secure cookies, real origins |
| prod | same-origin | SSL required (`DB_SSLMODE`) | private bucket + lifecycle policy | `APP_ENV=production` enables strict config checks |

Config keys live in `backend/.env.example` — validation fails fast on missing
values; `APP_ENV=production` additionally refuses weak session secrets,
`DB_SSLMODE=disable`, empty allowed origins, and insecure cookies.

## Testing

```bash
cd backend && go test ./internal/... ./tests/...   # unit + contract + integration
cd frontend && npm test -- --run && npm run lint && npm run build
cd frontend && npm run e2e:reset && npm run e2e  # Playwright critical path
```

Contract/integration tests spin a real Fiber app + Postgres schema; they need
`docker compose up -d` running and `TEST_DATABASE_URL` (defaults to
`postgres://postgres:postgres@127.0.0.1:5432/ojt_test`).

## Security posture

- Argon2id password hashing; opaque sessions hashed at rest, HttpOnly cookies.
- CSRF token required on all unsafe cookie-authenticated requests; Origin check.
- Role guards + coordinator trainee-scope enforced at the service layer
  (out-of-scope → 404, no existence leak).
- Client never supplies user/site/status/hours/watermark — server computes all.
- Login rate-limited; request body capped at the evidence upload limit.
- Private evidence bucket; originals immutable through application flows.

## Status

All five phases implemented; see `specs/001-ojt-attendance-daily-journal/tasks.md`.
Remaining unchecked items are institution-side (physical-device tests, UAT,
production deploy) — listed in `turnover.md`.

## Docs

- `docs/coordinator-guide.md` — staff workflows
- `docs/trainee-guide.md` — trainee quick reference
- `docs/deployment.md` — production deploy, backups, monitoring, rollback
- `docs/dependencies.md` — version + environment inventory
- `turnover.md` — handover: verified state, pending institution items, T136 deferral
