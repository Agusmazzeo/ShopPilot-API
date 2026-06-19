.PHONY: help migrate-up migrate-down migrate-status migrate-create

help:
	@echo "Available commands:"
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
