# Init the project

One repo, one backend process, one React app. Backend and frontend live in separate folders and talk over HTTP. Postgres and Redis run beside the API in local development. Kafka/RabbitMQ stays out of this step; the waiting room is a later feature inside the same codebase.

## Go framework

No. Do not adopt Gin, Echo, Fiber, Buffalo, Beego, or Revel.

The sale path is a transaction: lock or conditional update, hold expiry, and idempotent confirm. A framework owns routing, binding, and often the context type, and it hides the SQL that decides who gets the seat. That is the part we need to see.

Use the standard library for HTTP, and add libraries only where the standard library stops:

| Concern | Choice | Why |
| --- | --- | --- |
| HTTP | `net/http` ServeMux (Go 1.22+) | Method and path routing, middleware as plain wrappers |
| SQL | `pgx` + `sqlc` | Inventory updates stay as explicit SQL, including `UPDATE ... WHERE status = 'available'` |
| Migrations | `goose` | Versioned SQL next to the schema |
| Redis | `go-redis` | Holds and cache later; client only at init |
| Log | `log/slog` | Structured logs without a framework logger |
| Config | environment variables | One small loader in `internal/platform/config` |

`chi` is the only router worth adding later, and only if route groups become painful. Do not start with it.

Do not use an ORM (GORM, ent). A held or sold seat is a conditional write, not a mapped object save.

## Shape

Modular monolith:

- `backend/` is one Go module and one API binary, `cmd/api`.
- Domain code sits in `internal/` by area (events, inventory, orders), not by technical layer only.
- A later hold-expiry worker is a second `cmd/` in the same module, sharing `internal/`. That is still one system.
- `frontend/` is a Vite + React + TypeScript app. It is not a second backend. No Next.js.
- `deploy/docker-compose.yml` runs Postgres and Redis. The API and the UI run on the host during development.

## Layout

```text
myseat/
  backend/
    cmd/api/main.go
    internal/
      platform/config/
      platform/httpapi/      # router, middleware, health
      platform/postgres/
      platform/redis/
    migrations/
    go.mod
  frontend/                  # Vite React app
  deploy/docker-compose.yml
  docs/
```

Domain packages (`internal/event`, `internal/inventory`, `internal/order`) appear when those features start. Init does not stub empty business modules.

Existing Cursor React rules still point at `chat-app-frontend`. When the app exists, retarget those globs to `frontend/`.

## Local contract

- API: `http://localhost:8080`
- UI: Vite's default port
- Postgres: `localhost:5432`, database `myseat`
- Redis: `localhost:6379`
- Health: `GET /health` checks process, Postgres, and Redis
- Frontend calls the API with `VITE_API_URL`. Vite proxies `/api` to the backend in development so the browser stays same-origin.

Config comes from the environment. Copy `.env.example` files; do not commit real `.env` files.

## Phases

### Phase 1 — Repo skeleton

- Add the folders above, `backend/go.mod` (module `myseat`, current stable Go), and `frontend` via Vite's React TypeScript template.
- Add `.env.example` for the API and the UI.
- Extend the root `.gitignore` for Go build output, Vite `dist/`, and local env files.

### Phase 2 — API process

- `cmd/api` loads config, builds the router, and shuts down on SIGINT/SIGTERM.
- Routes: `GET /health`.
- JSON errors use one small response shape: `code`, `message`.

### Phase 3 — Postgres and Redis

- Compose file with Postgres 18 and Redis 7, healthchecks, and a named volume for Postgres.
- API opens both clients on startup and pings them from `/health`.
- Goose migration tool wired; first migration is empty or a `schema_migrations` sanity check only. No seat tables yet.

### Phase 4 — Frontend shell

- One page that requests `/health` and shows ok or failure.
- Root wrapper follows the project React rule: a single `div` with a kebab-case `*-wrapper` class.
- No event, seat, or checkout UI in this phase.

## Not in this init

- Auth, venues, events, holds, payments, tickets
- Kafka, RabbitMQ, or an in-process waiting room
- Seat-map cache
- Deploy to a real host, CI, or Kubernetes

After init, the next doc is the first vertical slice: create a venue and event, list events, and read a seat map from Postgres with no concurrency control yet.
