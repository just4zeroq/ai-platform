#!/bin/bash
# ============================================================
# migrate.sh — run goose migrations for all services
# Requires: goose CLI (go install github.com/pressly/goose/v3/cmd/goose@latest)
# Requires: PostgreSQL running (use: docker compose up -d postgres)
# ============================================================
set -e

cd "$(dirname "$0")/.."

DB_USER="${DB_USER:-aiplatform}"
DB_PASS="${DB_PASS:-aiplatform}"
DB_HOST="${DB_HOST:-localhost}"

run_migrations() {
  local svc="$1"
  local dbname="$2"
  echo "=== Migrating $svc ($dbname) ==="
  goose -dir "$svc/migrations" postgres "postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:5432/${dbname}?sslmode=disable" up
  echo ""
}

run_migrations user-svc  user_svc
run_migrations asset-svc asset_svc
run_migrations ai-gateway ai_gateway

echo "All migrations complete."
