#!/usr/bin/env bash
# Per-boot runtime initialization for the StarByte Cloud Agent environment.
# Brings up PostgreSQL, Redis and MinIO, provisions the dev database/role,
# applies migrations and seeds idempotent test data. Safe to run repeatedly.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DB_USER="starbyte"
DB_PASSWORD="starbyte"
DB_NAME="starbyte_dev"
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable"

echo "[start] Starting PostgreSQL..."
sudo pg_ctlcluster 16 main start 2>/dev/null || true
for _ in $(seq 1 30); do
  if pg_isready -h localhost -p 5432 >/dev/null 2>&1; then break; fi
  sleep 1
done

echo "[start] Ensuring database role and dev database exist..."
sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'" | grep -q 1 \
  || sudo -u postgres psql -c "CREATE ROLE ${DB_USER} LOGIN PASSWORD '${DB_PASSWORD}' CREATEDB;"
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1 \
  || sudo -u postgres createdb -O "${DB_USER}" "${DB_NAME}"

echo "[start] Starting Redis..."
redis-cli ping >/dev/null 2>&1 || sudo redis-server --daemonize yes --appendonly yes

echo "[start] Starting MinIO..."
if ! curl -sf http://localhost:9000/minio/health/live >/dev/null 2>&1; then
  mkdir -p /tmp/minio-data
  MINIO_ROOT_USER=minioadmin MINIO_ROOT_PASSWORD=minioadmin \
    nohup minio server /tmp/minio-data --address ":9000" --console-address ":9001" \
    > /tmp/minio.log 2>&1 &
fi
for _ in $(seq 1 30); do
  if curl -sf http://localhost:9000/minio/health/live >/dev/null 2>&1; then break; fi
  sleep 1
done
mc alias set local http://localhost:9000 minioadmin minioadmin >/dev/null 2>&1 || true
mc mb --ignore-existing local/starbyte >/dev/null 2>&1 || true

echo "[start] Applying database migrations..."
migrate -path "$REPO_ROOT/backend/migrations" -database "$DATABASE_URL" up

echo "[start] Seeding idempotent test data..."
cd "$REPO_ROOT/backend"
APP_ENV=dev DB_NAME="${DB_NAME}" go run ./scripts

echo "[start] Infrastructure ready (PostgreSQL, Redis, MinIO)."
