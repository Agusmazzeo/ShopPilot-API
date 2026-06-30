.PHONY: help test test-verbose test-coverage migrate-up migrate-down migrate-status migrate-create

help:
	@echo "Available commands:"
	@echo "  make test             - Run all tests"
	@echo "  make test-verbose     - Run all tests with verbose output"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-status   - Show migration status"
	@echo "  make migrate-create   - Create new migration (NAME=migration_name)"

migrate-up:
	@echo "Running migrations..."
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back migration..."
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	@echo "Migration status:"
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
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
