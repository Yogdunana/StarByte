#!/bin/bash
set -e

# ─────────────────────────────────────────────────────────
# StarByte 开发环境一键启动脚本
# 用法：./scripts/dev.sh
# ─────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info()    { echo -e "${CYAN}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC}   $1"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()   { echo -e "${RED}[FAIL]${NC} $1"; }

# ── 检查依赖 ──────────────────────────────────────────────
info "检查开发环境依赖..."

MISSING_DEPS=()

if ! command -v docker &> /dev/null; then
    MISSING_DEPS+=("Docker")
else
    success "Docker: $(docker --version | awk '{print $3}')"
fi

if ! command -v go &> /dev/null; then
    warn "Go: 未找到（可选，如果只用 Docker 开发可忽略）"
else
    success "Go: $(go version | awk '{print $3}')"
fi

if ! command -v node &> /dev/null; then
    warn "Node.js: 未找到（可选，如果只用 Docker 开发可忽略）"
else
    success "Node.js: $(node --version)"
fi

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    error "缺少必要依赖: ${MISSING_DEPS[*]}"
    echo "请先安装以上依赖后重试。"
    exit 1
fi

# ── 检查 docker compose 命令 ──────────────────────────────
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
elif docker-compose --version &> /dev/null; then
    COMPOSE_CMD="docker-compose"
else
    error "未找到 docker compose 或 docker-compose 命令"
    exit 1
fi
success "Compose: $($COMPOSE_CMD version --short)"

# ── 检查 .env 文件 ───────────────────────────────────────
ENV_FILE="$PROJECT_ROOT/deploy/.env"
if [ ! -f "$ENV_FILE" ]; then
    info "创建默认 .env 文件..."
    cp "$PROJECT_ROOT/deploy/.env.example" "$ENV_FILE"
    success "已创建 deploy/.env（如需自定义请编辑该文件）"
fi

# 从 .env 加载变量（只读取我们需要的，不污染全局）
load_env_var() {
    local key="$1"
    local default="$2"
    local value
    value=$(grep -E "^${key}=" "$ENV_FILE" 2>/dev/null | head -1 | cut -d'=' -f2- | tr -d '\r' || true)
    if [ -z "$value" ]; then
        echo "$default"
    else
        echo "$value"
    fi
}

POSTGRES_USER=$(load_env_var "POSTGRES_USER" "starbyte")
POSTGRES_PASSWORD=$(load_env_var "POSTGRES_PASSWORD" "starbyte")
POSTGRES_DB=$(load_env_var "POSTGRES_DB" "starbyte_dev")

# ── 启动基础设施 ──────────────────────────────────────────
echo ""
info "启动基础设施（PostgreSQL + Redis + MinIO）..."
$COMPOSE_CMD -f "$PROJECT_ROOT/deploy/docker-compose.dev.yml" --env-file "$ENV_FILE" up -d

# 等待健康检查 —— 不依赖 Python，直接用 docker inspect
info "等待基础设施就绪..."
MAX_WAIT=60
WAIT_COUNT=0
ALL_HEALTHY=false

check_all_healthy() {
    local containers
    containers=$($COMPOSE_CMD -f "$PROJECT_ROOT/deploy/docker-compose.dev.yml" ps -q 2>/dev/null)
    if [ -z "$containers" ]; then
        return 1
    fi
    for cid in $containers; do
        local health
        health=$(docker inspect --format='{{.State.Health.Status}}' "$cid" 2>/dev/null || echo "none")
        # 没有 healthcheck 的服务视为 healthy
        if [ "$health" != "none" ] && [ "$health" != "healthy" ]; then
            return 1
        fi
    done
    return 0
}

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if check_all_healthy; then
        ALL_HEALTHY=true
        break
    fi
    sleep 2
    WAIT_COUNT=$((WAIT_COUNT + 2))
    echo -n "."
done
echo ""

if [ "$ALL_HEALTHY" = true ]; then
    success "基础设施全部就绪"
else
    warn "等待超时，部分服务可能未就绪（继续执行）"
fi

# ── 运行数据库迁移 ────────────────────────────────────────
info "运行数据库迁移..."

if command -v migrate &> /dev/null; then
    migrate -path "$PROJECT_ROOT/backend/migrations" \
        -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" up
    success "数据库迁移完成"
else
    warn "未找到 migrate 命令，跳过自动迁移"
    echo "  安装命令: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo "  或手动运行: migrate -path backend/migrations -database 'postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable' up"
fi

# ── 启动后端（可选） ──────────────────────────────────────
echo ""
info "基础设施已就绪。"
echo ""
echo -e "${CYAN}后端启动:${NC}"
echo "  cd backend && go run cmd/server/main.go"
echo ""
echo -e "${CYAN}前端启动:${NC}"
echo "  cd frontend && npm install && npm run dev"
echo ""
echo -e "${CYAN}服务地址:${NC}"
echo "  前端:    http://localhost:5173"
echo "  后端API: http://localhost:8080/api/v1"
echo "  健康检查: http://localhost:8080/health"
echo "  MinIO:   http://localhost:9001 (minioadmin/minioadmin)"
echo ""
echo -e "${CYAN}停止基础设施:${NC}"
echo "  $COMPOSE_CMD -f deploy/docker-compose.dev.yml down"
echo ""
