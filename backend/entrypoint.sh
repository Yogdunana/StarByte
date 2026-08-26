#!/bin/bash
set -e

echo "[entrypoint] StarByte 后端服务启动中..."

# ── 从环境变量构建数据库连接串 ──────────────────────────────
DB_HOST="${DATABASE_HOST:-postgres}"
DB_PORT="${DATABASE_PORT:-5432}"
DB_USER="${DATABASE_USER:-starbyte}"
DB_PASSWORD="${DATABASE_PASSWORD:-starbyte}"
DB_NAME="${DATABASE_NAME:-starbyte}"
DB_SSLMODE="${DATABASE_SSLMODE:-disable}"

MIGRATE_DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# ── 自动执行数据库迁移 ────────────────────────────────────
if [ "${SKIP_MIGRATION:-false}" != "true" ]; then
    echo "[entrypoint] 执行数据库迁移..."
    if ! migrate -path /app/migrations -database "$MIGRATE_DSN" up; then
        echo "[entrypoint] ⚠️  数据库迁移失败，跳过（将继续启动服务）"
    else
        echo "[entrypoint] 数据库迁移完成"
    fi
else
    echo "[entrypoint] 跳过数据库迁移（SKIP_MIGRATION=true）"
fi

# ── 启动服务 ────────────────────────────────────────────
echo "[entrypoint] 启动 StarByte 服务..."
exec ./starbyte-server
