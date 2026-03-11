# snip.ly — URL Shortener

A production-ready URL shortening service built with Go, PostgreSQL, and Redis.
Supports admin authentication, click analytics, password reset via email, and a clean SaaS-style UI.

---

## Features

- **URL shortening** — 7-character nanoid short codes, duplicate detection, optional TTL
- **Fast redirects** — Redis cache-aside pattern; Postgres fallback with graceful degradation
- **Click analytics** — per-link daily breakdown with sparkbar visualisation
- **Admin dashboard** — protected route with session-based authentication
- **Account management** — registration with email activation, login, password reset flow
- **Email delivery** — async worker pool with SMTP (Mailtrap-compatible)
- **Rate limiting** — per-IP token bucket on all routes
- **URL sanitization** — blocks private IPs, loopback, unsafe schemes
- **Request timeouts** — configurable per-request context deadline
- **CORS** — configurable allowed origins for API consumers
- **Structured logging** — slog with JSON (production) and text (development) output
- **Health check** — `/health` endpoint with Postgres + Redis liveness
- **Graceful shutdown** — drains in-flight requests on SIGINT/SIGTERM
- **Docker** — multi-stage build, docker-compose with healthchecks
- **CI/CD** — GitHub Actions test pipeline + Docker Hub deploy workflow

---

## Tech Stack

| Layer      | Technology                     |
| ---------- | ------------------------------ |
| Language   | Go 1.23                        |
| Web        | Gin                            |
| Database   | PostgreSQL 16                  |
| Cache      | Redis 7                        |
| Sessions   | alexedwards/scs                |
| Migrations | golang-migrate                 |
| Email      | gopkg.in/mail.v2 + worker pool |
| Config     | Viper + .env                   |
| CSS        | Tailwind v4                    |
| Logging    | log/slog (stdlib)              |
| Container  | Docker + docker-compose        |

---

## Project Structure

```
cmd/web/         → entrypoint — wires all layers, starts server
internal/
  config/        → Viper-based typed config from .env
  handler/       → HTTP handlers (thin — delegate to service)
  service/       → business logic (no HTTP, no SQL)
  repository/    → database + cache interfaces and implementations
  middleware/    → rate limit, metrics, auth, CORS, timeout
  models/        → domain structs (URL, Admin, ClickEvent)
  mailer/        → SMTP client + async worker pool
  helpers/       → template rendering
pkg/
  logger/        → slog wrapper (JSON/text by env)
  sanitize/      → URL validation and normalisation
views/           → html/template layouts, pages, partials, emails
migrations/      → SQL up/down migration files
```

---

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 16
- Redis 7
- Node.js 20+ (for Tailwind CSS)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Local Setup

```bash
# Clone
git clone https://github.com/yourname/url-shortener.git
cd url-shortener

# Install Go dependencies
go mod download

# Install Node dependencies (Tailwind)
npm install

# Copy and configure environment
cp .env.example .env
# Edit .env with your DB, Redis, and mail credentials

# Create database
make db/create

# Run migrations
make migrate/up

# Build CSS
make css/build

# Start server
make run
```

Visit [http://localhost:8080](http://localhost:8080)

---

## Docker

```bash
# Start everything (app + postgres + redis + migrations)
docker compose up -d

# View logs
docker compose logs -f app

# Tear down
docker compose down
```

---

## Environment Variables

| Variable                | Default                                               | Description                        |
| ----------------------- | ----------------------------------------------------- | ---------------------------------- |
| `ENV`                   | `development`                                         | `development` or `production`      |
| `APP_BASE_URL`          | `http://localhost:8080`                               | Base URL for generated short links |
| `SERVER_PORT`           | `8080`                                                | HTTP server port                   |
| `DB_DSN`                | `postgres://urluser:urlpassword@localhost:5432/urldb` | PostgreSQL connection string       |
| `DB_MAX_OPEN_CONNS`     | `25`                                                  | Max open DB connections            |
| `DB_MAX_IDLE_CONNS`     | `5`                                                   | Max idle DB connections            |
| `DB_CONN_MAX_LIFETIME`  | `5m`                                                  | Max connection lifetime            |
| `REDIS_ADDR`            | `localhost:6379`                                      | Redis address                      |
| `REDIS_CACHE_TTL`       | `24h`                                                 | Redis cache TTL                    |
| `SESSION_LIFETIME`      | `12h`                                                 | Admin session duration             |
| `SESSION_SECURE_COOKIE` | `false`                                               | Set `true` behind HTTPS            |
| `MAIL_HOST`             | `sandbox.smtp.mailtrap.io`                            | SMTP host                          |
| `MAIL_PORT`             | `587`                                                 | SMTP port                          |
| `MAIL_USERNAME`         | —                                                     | SMTP username                      |
| `MAIL_PASSWORD`         | —                                                     | SMTP password                      |
| `MAIL_FROM`             | `noreply@snip.ly`                                     | Sender address                     |
| `MAIL_WORKERS`          | `3`                                                   | Email worker goroutines            |
| `CORS_ALLOWED_ORIGINS`  | `http://localhost:3000`                               | Comma-separated allowed origins    |

---

## Make Targets

```bash
make run              # start the Go server
make dev              # start Go + Tailwind watch concurrently
make build            # compile to ./bin/url
make css/build        # build Tailwind CSS once
make test             # run all tests
make test/verbose     # run tests with -v output
make vet              # go vet
make migrate/up       # apply all pending migrations
make migrate/down     # roll back last migration
make migrate/version  # show current version
make docker/build     # build Docker image
make docker/up        # start all Docker services
make docker/down      # stop all Docker services
make docker/logs      # tail app container logs
```

---

## API

The shortener also responds to JSON clients:

### `POST /shorten`

**Request**

```json
{ "url": "https://example.com/very/long/path" }
```

**Response `200`**

```json
{
  "short_code": "hygR_gx",
  "short_url": "https://snip.ly/hygR_gx"
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

## CI/CD

| Workflow     | Trigger             | What it does                                  |
| ------------ | ------------------- | --------------------------------------------- |
| `test.yml`   | push / pull request | `go vet` + `go test -race` with real services |
| `deploy.yml` | push to `main`      | Build image → push to Docker Hub → SSH deploy |

Required GitHub secrets: `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`

---

## Architecture

```
Browser / API Client
       │
  routes.go          ← all routes registered here
       │
  middleware          ← CORS → Timeout → RateLimit → Session → Auth
       │
  handler/            ← HTTP in, HTTP out — thin layer
       │
  service/            ← all business logic, no HTTP, no SQL
       │
  repository/         ← interfaces → Postgres (source of truth)
                                   → Redis    (cache layer)
```

**Cache-aside pattern:** every redirect checks Redis first. On miss, queries Postgres and writes back to Redis. Redis failures are non-fatal — the service degrades gracefully to Postgres.

---

## License

MIT

```

---
```
