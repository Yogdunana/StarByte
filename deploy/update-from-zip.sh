#!/usr/bin/env bash
# 用 zip 覆盖更新 StarByte 代码。
# - 始终保留现有 deploy/.env（不覆盖密钥）
# - 绝不执行 docker volume / compose down -v
set -euo pipefail

usage() {
  cat <<'EOF'
用法: bash deploy/update-from-zip.sh <StarByte.zip>

将 zip 解压并覆盖到本仓库根目录：
  • 保留已有 deploy/.env
  • 不触碰 Docker named volumes（postgres/redis/minio 数据）
  • 不删除 .git
  • deploy/docker-compose.yml 会被 zip 覆盖（CAS extra_hosts 已在仓库 compose 内）

zip 可以是 GitHub 的 Source code（顶层带 StarByte-main/ 目录），
也可以是仓库根文件直接打包。

更新后请重建应用（不会强拉 MinIO）：
  starbyte rebuild all
  # 或
  docker compose -f deploy/docker-compose.yml up -d --build --no-deps --force-recreate backend frontend
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

ZIP="${1:-}"
if [[ -z "$ZIP" ]]; then
  usage >&2
  exit 2
fi
if [[ ! -f "$ZIP" ]]; then
  echo "找不到 zip: $ZIP" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$REPO_ROOT/deploy/docker-compose.yml"
ENV_FILE="$REPO_ROOT/deploy/.env"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "未找到 $COMPOSE_FILE，请在 StarByte 仓库内执行本脚本。" >&2
  exit 1
fi

WORKDIR="$(mktemp -d)"
cleanup() { rm -rf "$WORKDIR"; }
trap cleanup EXIT

if [[ -f "$ENV_FILE" ]]; then
  cp -a "$ENV_FILE" "$WORKDIR/preserved.env"
  echo "已暂存 deploy/.env"
fi

EXTRACT="$WORKDIR/extract"
mkdir -p "$EXTRACT"

if command -v unzip >/dev/null 2>&1; then
  unzip -q "$ZIP" -d "$EXTRACT"
elif command -v python3 >/dev/null 2>&1; then
  python3 -m zipfile -e "$ZIP" "$EXTRACT"
else
  echo "需要 unzip 或 python3 才能解压。" >&2
  exit 1
fi

# 顶层若只有一个目录，视为 GitHub 源码包
shopt -s nullglob
entries=("$EXTRACT"/*)
shopt -u nullglob
SRC="$EXTRACT"
if [[ ${#entries[@]} -eq 1 && -d "${entries[0]}" ]]; then
  SRC="${entries[0]}"
fi

if [[ ! -f "$SRC/deploy/docker-compose.yml" && ! -f "$SRC/README.md" ]]; then
  echo "zip 内容不像 StarByte 仓库根（缺少 deploy/docker-compose.yml 或 README.md）。" >&2
  exit 1
fi

# 用 tar 覆盖文件；排除 .git 与现有 .env（稍后再还原）
# 不使用 rsync，减少校园机依赖
echo "正在覆盖: $REPO_ROOT  （来源: $SRC）"
tar -C "$SRC" --exclude='.git' --exclude='deploy/.env' -cf - . \
  | tar -C "$REPO_ROOT" -xf -

if [[ -f "$WORKDIR/preserved.env" ]]; then
  mkdir -p "$(dirname "$ENV_FILE")"
  cp -a "$WORKDIR/preserved.env" "$ENV_FILE"
  echo "已还原 deploy/.env"
else
  echo "目录中原先没有 deploy/.env（未新建密钥文件）"
fi

if ! grep -q 'extra_hosts' "$COMPOSE_FILE" \
  || ! grep -q 'authserver.smbu.edu.cn' "$COMPOSE_FILE"; then
  echo "警告: 更新后的 deploy/docker-compose.yml 没有 authserver extra_hosts。" >&2
  echo "  CAS callback 可能再解析到 IPv6 并 502。请换含本仓库 compose 的 zip，或手工加回 extra_hosts。" >&2
fi

cat <<EOF

代码已更新。Docker named volumes 未被本脚本触碰。
deploy/docker-compose.yml 已按 zip 覆盖（CAS extra_hosts 应已在文件内）。
下一步（按需）：
  starbyte rebuild all
  # 或 docker compose -f deploy/docker-compose.yml up -d --build --no-deps --force-recreate backend frontend

切勿: docker compose down -v   （会删 postgres/redis/minio 数据）
EOF
