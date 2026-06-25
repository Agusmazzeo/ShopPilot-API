#!/bin/bash
set -e

echo "Testing database connection..."
cd "$(dirname "$0")/.."
source settings/envVars.LOCAL

echo "✓ Environment loaded"
echo "  DATABASE_URL: ${DATABASE_URL}"

echo ""
echo "Testing PostgreSQL connection..."
psql "$DATABASE_URL" -c "SELECT version();"

echo ""
echo "Testing goose installation..."
which goose || echo "ERROR: goose not installed. Run: go install github.com/pressly/goose/v3/cmd/goose@latest"

echo ""
echo "Current migration status..."
goose -dir migrations postgres "$DATABASE_URL" status || echo "No migrations applied yet"

echo ""
echo "✓ All checks passed"
