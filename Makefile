# ==================================================================================
# URL Shortener Makefile
# ==================================================================================

# Variables
DB_URL=postgres://urluser:urlpassword@localhost:5432/urldb?sslmode=disable
MIGRATIONS_PATH=migrations
CMD_PATH=./cmd/web

# ==================================================================================
# Development
# ==================================================================================

## run: start the Go server
run:
	go run $(CMD_PATH)

## dev: start Go server + Tailwind watch concurrently
dev:
	npm run dev

## css/build: build Tailwind CSS once
css/build:
	npm run css:build

## css/watch: watch and rebuild Tailwind CSS on change
css/watch:
	npm run css:watch

# ==================================================================================
# Build
# ==================================================================================

## build: compile the app into ./bin/url
build:
	go build -o ./bin/url $(CMD_PATH)

## build/css: build CSS then compile Go app
build/all: css/build build

# ==================================================================================
# Database / Migrations
# ==================================================================================

## db/create: create the database and user (run once)
db/create:
	psql -U postgres -c "CREATE USER urluser WITH PASSWORD 'urlpassword';"
	psql -U postgres -c "CREATE DATABASE urldb OWNER urluser;"
	psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE urldb TO urluser;"

## migrate/up: apply all pending migrations
migrate/up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

## migrate/down: roll back the last migration
migrate/down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

## migrate/down/all: roll back all migrations
migrate/down/all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

## migrate/version: show current migration version
migrate/version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

## migrate/force: force a specific version (usage: make migrate/force v=1)
migrate/force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(v)

# ==================================================================================
# Redis
# ==================================================================================

## redis/start: start Redis service
redis/start:
	sc start Redis

## redis/stop: stop Redis service  
redis/stop:
	sc stop Redis

## redis/ping: check Redis is running
redis/ping:
	"C:/Program Files/Redis/redis-cli.exe" ping



# ==================================================================================
# Go tooling
# ==================================================================================

## tidy: tidy go modules
tidy:
	go mod tidy

## test: run all tests
test:
	go test ./...

## test/verbose: run all tests with verbose output
test/verbose:
	go test -v ./...

## vet: run go vet
vet:
	go vet ./...


# ==================================================================================
# Docker
# ==================================================================================
## docker/build: build the Docker image
docker/build:
	docker compose build

## docker/up: start all services
docker/up:
	docker compose up -d

## docker/down: stop all services
docker/down:
	docker compose down

## docker/logs: tail app logs
docker/logs:
	docker compose logs -f app

## docker/migrate: run migrations inside Docker
docker/migrate:
	docker compose run --rm migrate	

# ==================================================================================
# Help
# ==================================================================================

## help: print this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: run dev css/build css/watch build build/all db/create \
	migrate/up migrate/down migrate/down/all migrate/version migrate/force \
	tidy test test/verbose vet help