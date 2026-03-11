# snip.ly

> A production-ready URL shortening service built with Go, PostgreSQL, and Redis.

---

## Features

- **URL shortening** — 7-character nanoid short codes with duplicate detection
- **Fast redirects** — Redis cache-aside; Postgres fallback with graceful degradation
- **Click analytics** — per-link daily breakdown with sparkbar visualisation
- **Admin dashboard** — protected route with session-based authentication
- **Password reset** — secure email-based flow with time-limited tokens
- **Rate limiting** — per-IP token bucket on all public routes
- **URL sanitization** — blocks private IPs, loopback addresses, and unsafe schemes
- **Request timeouts** — per-request context cancellation deadline
- **CORS** — configurable allowed origins for API consumers
- **Structured logging** — JSON in production, text in development (stdlib slog)
- **Health check** — `/health` with Postgres + Redis liveness
- **Graceful shutdown** — drains in-flight requests on SIGINT/SIGTERM
- **Dockerized** — multi-stage build with docker-compose and healthchecks
- **CI/CD** — GitHub Actions test pipeline and Docker Hub deploy workflow

---

## Tech Stack

| Layer            | Technology        |
| ---------------- | ----------------- |
| Language         | Go 1.23           |
| Web Framework    | Gin               |
| Database         | PostgreSQL 16     |
| Cache            | Redis 7           |
| Migrations       | golang-migrate    |
| Sessions         | alexedwards/scs   |
| Email            | gopkg.in/mail.v2  |
| Config           | Viper             |
| CSS              | Tailwind v4       |
| Logging          | log/slog (stdlib) |
| Containerization | Docker            |
| CI/CD            | GitHub Actions    |

---

## Project Structure

```
cmd/web/             → entrypoint — wires all layers, starts server
internal/
  config/            → Viper-based typed config from .env
  handler/           → HTTP handlers (thin — delegate to service)
  service/           → business logic (no HTTP, no SQL)
  repository/        → database and cache interfaces + implementations
  middleware/        → rate limit, metrics, auth, CORS, timeout
  models/            → domain structs (URL, Admin, ClickEvent)
  mailer/            → SMTP client + async worker pool
  helpers/           → template rendering
pkg/
  logger/            → slog wrapper (JSON/text by env)
  sanitize/          → URL validation and normalisation
views/               → html/template layouts, pages, partials, emails
migrations/          → SQL up/down migration files
```

---

## Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 16
- Redis 7
- Node.js 20+ (Tailwind CSS)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Run locally

```bash
git clone https://github.com/yourname/snip.ly.git
cd snip.ly

go mod download
npm install

cp .env.example .env
# edit .env with your credentials

make db/create
make migrate/up
make css/build
make run
```

Visit `http://localhost:8080`

### Run with Docker

```bash
docker compose up -d
```

This starts the app, Postgres, Redis, and runs migrations automatically.

---

## Environment Variables

| Variable                | Default                                               | Description                        |
| ----------------------- | ----------------------------------------------------- | ---------------------------------- |
| `ENV`                   | `development`                                         | `development` or `production`      |
| `APP_BASE_URL`          | `http://localhost:8080`                               | Base URL for generated short links |
| `SERVER_PORT`           | `8080`                                                | HTTP listen port                   |
| `DB_DSN`                | `postgres://urluser:urlpassword@localhost:5432/urldb` | PostgreSQL connection string       |
| `DB_MAX_OPEN_CONNS`     | `25`                                                  | Max open DB connections            |
| `DB_MAX_IDLE_CONNS`     | `5`                                                   | Max idle DB connections            |
| `DB_CONN_MAX_LIFETIME`  | `5m`                                                  | Max connection lifetime            |
| `REDIS_ADDR`            | `localhost:6379`                                      | Redis address                      |
| `REDIS_CACHE_TTL`       | `24h`                                                 | Redis cache TTL per entry          |
| `SESSION_LIFETIME`      | `12h`                                                 | Admin session duration             |
| `SESSION_SECURE_COOKIE` | `false`                                               | Set `true` when behind HTTPS       |
| `MAIL_HOST`             | `sandbox.smtp.mailtrap.io`                            | SMTP host                          |
| `MAIL_PORT`             | `587`                                                 | SMTP port                          |
| `MAIL_USERNAME`         | —                                                     | SMTP username                      |
| `MAIL_PASSWORD`         | —                                                     | SMTP password                      |
| `MAIL_FROM`             | `noreply@snip.ly`                                     | Sender address                     |
| `MAIL_WORKERS`          | `3`                                                   | Email worker goroutines            |
| `CORS_ALLOWED_ORIGINS`  | `http://localhost:3000`                               | Comma-separated allowed origins    |

---

## API

### `POST /shorten`

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/path"}'
```

```json
{
  "short_code": "hygR_gx",
  "short_url": "http://localhost:8080/hygR_gx"
}
```

### `GET /:code`

Redirects to the original URL with `301 Moved Permanently`.

### `GET /:code/stats`

Returns click analytics for a short code.

### `GET /health`

```json
{
  "status": "ok",
  "checks": {
    "postgres": "ok",
    "redis": "ok"
  }
}
```

---

## Make Targets

```bash
make run              # start the Go server
make dev              # start Go + Tailwind watch concurrently
make build            # compile binary to ./bin/url
make css/build        # build Tailwind CSS once
make test             # run all tests
make test/verbose     # run tests with -v
make vet              # go vet
make migrate/up       # apply all pending migrations
make migrate/down     # roll back last migration
make migrate/version  # show current migration version
make docker/build     # build Docker image
make docker/up        # start all Docker services
make docker/down      # stop all Docker services
make docker/logs      # tail app container logs
```

---

## Architecture

```
Browser / API Client
        │
   routes.go              ← all routes registered here
        │
   middleware              ← CORS → Timeout → RateLimit → Session → Auth
        │
   handler/                ← HTTP in, HTTP out — thin layer
        │
   service/                ← all business logic, no HTTP, no SQL
        │
   repository/             ← interfaces
        ├── PostgresStore  ← source of truth
        └── RedisCache     ← fast cache layer (cache-aside)
```

**Redirect flow:** every `GET /:code` checks Redis first. On miss, queries Postgres and writes back to Redis. Redis failures are non-fatal — the service degrades transparently to Postgres-only.

---

## CI/CD

| Workflow     | Trigger             | What it does                                  |
| ------------ | ------------------- | --------------------------------------------- |
| `test.yml`   | push / pull request | `go vet` + `go test -race` with real services |
| `deploy.yml` | push to `main`      | Build image → push to Docker Hub → SSH deploy |

Required secrets: `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`

---

## License

MIT
