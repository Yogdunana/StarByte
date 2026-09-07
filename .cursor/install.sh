#!/usr/bin/env bash
# Idempotent repository bootstrap for the StarByte Cloud Agent environment.
# Installs system services (PostgreSQL 16, Redis, MinIO) and CLI tooling
# (golang-migrate, mc), then fetches backend/frontend dependencies and warms
# the Go build cache. Runtime services are started in start.sh, not here.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATE_VERSION="v4.17.1"

echo "[install] Installing system packages (PostgreSQL 16, Redis, tooling)..."
export DEBIAN_FRONTEND=noninteractive
sudo apt-get update -qq
sudo apt-get install -y -qq \
  postgresql postgresql-contrib \
  redis-server \
  ca-certificates wget curl

echo "[install] Installing golang-migrate..."
if ! command -v migrate >/dev/null 2>&1; then
  wget -q "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" -O /tmp/migrate.tar.gz
  tar -xzf /tmp/migrate.tar.gz -C /tmp migrate
  sudo mv /tmp/migrate /usr/local/bin/migrate
  sudo chmod +x /usr/local/bin/migrate
fi

echo "[install] Installing MinIO server + client..."
if ! command -v minio >/dev/null 2>&1; then
  wget -q "https://dl.min.io/server/minio/release/linux-amd64/minio" -O /tmp/minio
  sudo mv /tmp/minio /usr/local/bin/minio
  sudo chmod +x /usr/local/bin/minio
fi
if ! command -v mc >/dev/null 2>&1; then
  wget -q "https://dl.min.io/client/mc/release/linux-amd64/mc" -O /tmp/mc
  sudo mv /tmp/mc /usr/local/bin/mc
  sudo chmod +x /usr/local/bin/mc
fi

echo "[install] Downloading Go module dependencies..."
cd "$REPO_ROOT/backend"
go mod download
# Warm the build cache so the first backend boot is fast.
go build -o /tmp/starbyte-server ./cmd/server

echo "[install] Installing frontend dependencies..."
cd "$REPO_ROOT/frontend"
npm ci

echo "[install] Done."
