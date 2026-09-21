#!/bin/bash
set -e

echo "[entrypoint] StarByte 后端服务启动中..."

# ── 默认按生产处理（fail-closed），只有显式 dev/test 才放宽 ──────────
# 本镜像可能被单独 `docker run`（不经 compose）。下面的 `${VAR:-默认}` 兜底
# 会让服务带着开发口令起起来，所以默认（未设置 APP_ENV）就按生产严格校验：
# 任一密钥缺失/为空就直接退出，而不是静默退化成 dev 口令。
# 只有显式声明 APP_ENV=dev 或 APP_ENV=test 时才走宽松分支。
# 正常 compose 部署由 starbyte init 写入 deploy/.env，这些变量一定存在。
case "${APP_ENV:-prod}" in
  dev|test) : ;;            # 显式声明才放宽
  *)
    : "${DB_PASSWORD:?生产环境必须设置 DB_PASSWORD（请运行 starbyte init 生成随机口令）}"
    : "${REDIS_PASSWORD:?生产环境必须设置 REDIS_PASSWORD（请运行 starbyte init 生成随机口令）}"
    : "${MINIO_ACCESS_KEY:?生产环境必须设置 MINIO_ACCESS_KEY（请运行 starbyte init 生成随机账号）}"
    : "${MINIO_SECRET_KEY:?生产环境必须设置 MINIO_SECRET_KEY（请运行 starbyte init 生成随机口令）}"
    : "${JWT_SECRET:?生产环境必须设置 JWT_SECRET（请运行 starbyte init 生成随机密钥）}"
    echo "[entrypoint] 生产环境密钥校验通过（值不打印）"
    ;;
esac

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

# 生产镜像内 mc 可能是占位脚本（不连 dl.min.io）。内网请设 SKIP_BUCKET_CREATE=true，
# 并在 MinIO 控制台手动创建桶。SKIP_BUCKET_CREATE=false 时若 mc 为 stub 会立刻成功返回。
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
    elif echo "$MB_OUTPUT" | grep -qi "already exists\|bucket.*exist\|already own\|previous request"; then
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
