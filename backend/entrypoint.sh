#!/bin/bash
set -e

echo "[entrypoint] StarByte 后端服务启动中..."

# ── 从环境变量构建数据库连接串 ──────────────────────────────
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-starbyte}"
DB_PASSWORD="${DB_PASSWORD:-starbyte}"
DB_NAME="${DB_NAME:-starbyte}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

MIGRATE_DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# ── 自动执行数据库迁移 ────────────────────────────────────
if [ "${SKIP_MIGRATION:-false}" != "true" ]; then
    echo "[entrypoint] 执行数据库迁移..."
    if ! migrate -path /app/migrations -database "$MIGRATE_DSN" up; then
        if [ "${MIGRATION_FAIL_FATAL:-true}" = "true" ]; then
            echo "[entrypoint] 数据库迁移失败，MIGRATION_FAIL_FATAL=true，退出容器"
            exit 1
        else
            echo "[entrypoint] 数据库迁移失败，跳过（将继续启动服务）"
        fi
    else
        echo "[entrypoint] 数据库迁移完成"
    fi
else
    echo "[entrypoint] 跳过数据库迁移（SKIP_MIGRATION=true）"
fi

# ── 自动创建 MinIO Bucket ─────────────────────────────────
MINIO_ENDPOINT="${MINIO_ENDPOINT:-minio:9000}"
MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-minioadmin}"
MINIO_BUCKET="${MINIO_BUCKET:-starbyte}"
MINIO_USE_SSL="${MINIO_USE_SSL:-false}"

if [ "${SKIP_BUCKET_CREATE:-false}" != "true" ]; then
    echo "[entrypoint] 检查 MinIO Bucket: ${MINIO_BUCKET}"
    SCHEME="http"
    if [ "$MINIO_USE_SSL" = "true" ]; then
        SCHEME="https"
    fi

    # 配置 mc alias（仅本地写配置，不验证连接）
    mc alias set starbyte "${SCHEME}://${MINIO_ENDPOINT}" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" > /dev/null 2>&1 || true

    # 直接尝试创建 bucket：已存在也算成功，其他错误打印详情
    if MB_OUTPUT=$(mc mb "starbyte/${MINIO_BUCKET}" 2>&1); then
        echo "[entrypoint] MinIO Bucket ${MINIO_BUCKET} 已创建"
    elif echo "$MB_OUTPUT" | grep -qi "already exists\|bucket.*exist"; then
        echo "[entrypoint] MinIO Bucket ${MINIO_BUCKET} 已存在"
    else
        echo "[entrypoint] MinIO Bucket 检查/创建失败，将继续启动服务"
        echo "  详情: $MB_OUTPUT"
    fi
else
    echo "[entrypoint] 跳过 MinIO Bucket 创建（SKIP_BUCKET_CREATE=true）"
fi

# ── 启动服务 ────────────────────────────────────────────
echo "[entrypoint] 启动 StarByte 服务..."
exec ./starbyte-server
