.PHONY: help test test-verbose test-coverage db-migrate-up db-migrate-down db-migrate-status db-migrate-create dc-db-up dc-db-down dc-redis-up dc-redis-down

help:
	@echo "Available commands:"
	@echo "  make test             - Run all tests"
	@echo "  make test-verbose     - Run all tests with verbose output"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make db-migrate-up       - Run all pending migrations"
	@echo "  make db-migrate-down     - Rollback last migration"
	@echo "  make db-migrate-status   - Show migration status"
	@echo "  make db-migrate-create - Create new migration (NAME=migration_name)"
	@echo "  make dc-db-up         - Start PostgreSQL container"
	@echo "  make dc-db-down       - Stop PostgreSQL container"
	@echo "  make dc-redis-up      - Start Redis container"
	@echo "  make dc-redis-down    - Stop Redis container"

db-migrate-up:
	@echo "Running migrations..."
	goose -dir migrations postgres "$(DATABASE_URL)" up

db-migrate-down:
	@echo "Rolling back migration..."
	goose -dir migrations postgres "$(DATABASE_URL)" down

db-migrate-status:
	@echo "Migration status:"
	goose -dir migrations postgres "$(DATABASE_URL)" status

db-migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make db-migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	goose -dir migrations create $(NAME) sql

test:
	@echo "Running tests..."
	go test ./...

test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	@echo "\nCoverage summary:"
	go tool cover -func=coverage.out

dc-db-up:
	@echo "Starting PostgreSQL container..."
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "Running migrations..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs) && goose -dir migrations postgres "$$DATABASE_URL" up; \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@echo "Database is ready!"

dc-db-down:
	@echo "Stopping PostgreSQL container..."
	docker-compose stop postgres
	docker-compose rm -f postgres
	docker volume rm shoppilot_postgres_data

dc-redis-up:
	@echo "Starting Redis container..."
	docker-compose up -d redis

dc-redis-down:
	@echo "Stopping Redis container..."
	docker-compose stop redis
	docker volume rm shoppilot_redis_data

dc-down:
	@echo "Stopping all containers..."
	make dc-db-down
	make dc-redis-down

dc-up:
	@echo "Starting all containers..."
	make dc-db-up
	make dc-redis-up