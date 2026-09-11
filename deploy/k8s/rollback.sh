#!/usr/bin/env bash
# 回滚 backend / frontend（kubectl rollout undo）
#
# 用法:
#   ./rollback.sh backend              # 或 starbyte-backend
#   ./rollback.sh frontend             # 或 starbyte-frontend
#   ./rollback.sh all
#   ./rollback.sh backend starbyte     # 第二参数为 namespace
set -euo pipefail

TARGET="${1:-backend}"
NAMESPACE="${2:-${STARBYTE_NAMESPACE:-starbyte}}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

resolve_name() {
  case "$1" in
    backend|starbyte-backend) echo "starbyte-backend" ;;
    frontend|starbyte-frontend) echo "starbyte-frontend" ;;
    all) echo "all" ;;
    *)
      echo -e "${RED}错误: 仅支持 backend | frontend | all | starbyte-backend | starbyte-frontend${NC}" >&2
      echo "用法: $0 {backend|frontend|all} [namespace]" >&2
      exit 1
      ;;
  esac
}

rollback_one() {
  local name="$1"
  echo -e "${YELLOW}=== 回滚 ${name} (ns=${NAMESPACE}) ===${NC}"
  kubectl rollout history "deployment/${name}" -n "${NAMESPACE}"
  kubectl rollout undo "deployment/${name}" -n "${NAMESPACE}"
  kubectl rollout status "deployment/${name}" -n "${NAMESPACE}" --timeout=5m
  echo -e "${GREEN}✅ ${name} 回滚完成${NC}"
}

if ! command -v kubectl >/dev/null 2>&1; then
  echo -e "${RED}错误: 未检测到 kubectl${NC}" >&2
  exit 1
fi

RESOLVED="$(resolve_name "${TARGET}")"
if [[ "${RESOLVED}" == "all" ]]; then
  rollback_one starbyte-backend
  rollback_one starbyte-frontend
else
  rollback_one "${RESOLVED}"
fi
