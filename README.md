snip.ly — URL Shortener

A production-ready URL shortening service built with Go, PostgreSQL, and Redis.

Supports URL shortening, admin authentication, click analytics, and fast redirects using Redis caching.

Features

URL shortening with 7-character nanoid codes

Fast redirects with Redis caching

Click analytics per short link

Admin dashboard with authentication

Email-based password reset

Rate limiting for abuse protection

URL sanitization for security

Health check endpoint

Dockerized deployment

CI/CD with GitHub Actions

Tech Stack
Layer	Technology
Language	Go
Web Framework	Gin
Database	PostgreSQL
Cache	Redis
Migrations	golang-migrate
Sessions	scs
Logging	slog
Containerization	Docker
CI/CD	GitHub Actions
Project Structure
cmd/web/          application entrypoint
internal/
  config/         application configuration
  handler/        HTTP handlers
  service/        business logic
  repository/     database and cache access
  middleware/     HTTP middleware
  models/         domain models
  mailer/         email worker
pkg/
  logger/         structured logging
  sanitize/       URL validation
views/            HTML templates
migrations/       database migrations
Quick Start
Run with Docker

MIT
