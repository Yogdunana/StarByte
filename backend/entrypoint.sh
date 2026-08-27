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

    # 配置 mc alias
    if mc alias set starbyte "${SCHEME}://${MINIO_ENDPOINT}" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" > /dev/null 2>&1; then
        # 检查 bucket 是否存在
        if mc ls "starbyte/${MINIO_BUCKET}" > /dev/null 2>&1; then
            echo "[entrypoint] MinIO Bucket ${MINIO_BUCKET} 已存在"
        else
            # 创建 bucket
            if mc mb "starbyte/${MINIO_BUCKET}" > /dev/null 2>&1; then
                echo "[entrypoint] MinIO Bucket ${MINIO_BUCKET} 已创建"
            else
                echo "[entrypoint] MinIO Bucket 创建失败，将继续启动"
            fi
        fi
    else
        echo "[entrypoint] MinIO 连接失败，跳过 Bucket 检查（将继续启动服务）"
    fi
else
    echo "[entrypoint] 跳过 MinIO Bucket 创建（SKIP_BUCKET_CREATE=true）"
fi

# ── 启动服务 ────────────────────────────────────────────
echo "[entrypoint] 启动 StarByte 服务..."
exec ./starbyte-server
