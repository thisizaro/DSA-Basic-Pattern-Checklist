# Common development commands. Run `make help` to see this list.

.PHONY: help run build migrate tidy fmt

help:
	@echo "Available commands:"
	@echo "  make run       - Run the server locally (reads .env)"
	@echo "  make build     - Build a production binary into ./bin/server"
	@echo "  make migrate   - Apply all migrations in order (needs DATABASE_URL env var)"
	@echo "  make tidy      - Tidy go.mod/go.sum"
	@echo "  make fmt       - Format all Go source files"

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

migrate:
	psql "$$DATABASE_URL" -f migrations/0001_init.sql
	psql "$$DATABASE_URL" -f migrations/0002_seed.sql
	psql "$$DATABASE_URL" -f migrations/0003_auth_and_profiles.sql

tidy:
	go mod tidy

fmt:
	gofmt -w .
