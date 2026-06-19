# ShopPilot API

Backend API for ShopPilot e-commerce platform.

## Tech Stack

- Go 1.25.11
- PostgreSQL 15+
- Redis 7+
- pgx/v5 for database
- goose for migrations
- gorilla/mux for routing

## Setup

### Prerequisites

- Go 1.25.11
- PostgreSQL 15+
- Redis 7+

### Environment Configuration

Copy the environment template:

```bash
cp settings/envVars.LOCAL settings/envVars.LOCAL.user
# Edit settings/envVars.LOCAL.user with your values
source settings/envVars.LOCAL.user
```

### Database Setup

Run migrations:

```bash
make migrate-up
```

### Running the Server

```bash
go run main.go
```

Or build and run:

```bash
go build -o bin/shoppilot .
./bin/shoppilot
```

## API Endpoints

### Health Checks

- `GET /health` - Overall system health
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

### API v1

- `GET /api/v1/...` - API endpoints (TBD)

## Development

### Running Tests

```bash
go test ./...
```

### Creating Migrations

```bash
make migrate-create NAME=your_migration_name
```
