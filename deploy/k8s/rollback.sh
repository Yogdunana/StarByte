#!/bin/bash
# StarByte Kubernetes 部署清单 - Issue #102
# 回滚脚本：使用 kubectl rollout undo 回滚 backend / frontend
#
# 用法: ./rollback.sh [deployment-name] [namespace]
#   deployment-name: 可选，默认 starbyte-backend（也可传 starbyte-frontend）
#   namespace:        可选，默认 starbyte
#
# 示例:
#   ./rollback.sh                          # 回滚 starbyte-backend 到上一版本
#   ./rollback.sh starbyte-frontend         # 回滚 starbyte-frontend
#   ./rollback.sh starbyte-backend prod      # 在 prod 命名空间回滚 backend
#
# 如需回滚到指定版本，可先查看下方历史版本，再手动执行：
#   kubectl rollout undo deployment/starbyte-backend -n starbyte --to-revision=<N>

set -euo pipefail

DEPLOYMENT_NAME="${1:-starbyte-backend}"
NAMESPACE="${2:-starbyte}"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== StarByte 回滚工具 ===${NC}"
echo -e "Deployment: ${GREEN}${DEPLOYMENT_NAME}${NC}"
echo -e "Namespace:  ${GREEN}${NAMESPACE}${NC}"
echo ""

# 依赖检查
if ! command -v kubectl >/dev/null 2>&1; then
  echo -e "${RED}错误: 未检测到 kubectl，请先安装并配置 kubeconfig。${NC}"
  exit 1
fi

# 仅允许回滚 backend / frontend，防止误操作数据类工作负载
if [[ "${DEPLOYMENT_NAME}" != "starbyte-backend" && "${DEPLOYMENT_NAME}" != "starbyte-frontend" ]]; then
  echo -e "${RED}错误: 仅支持回滚 starbyte-backend 或 starbyte-frontend，当前传入: ${DEPLOYMENT_NAME}${NC}"
  echo -e "可用回滚目标: starbyte-backend | starbyte-frontend"
  exit 1
fi

echo -e "${YELLOW}[1/3] 当前 rollout 状态:${NC}"
kubectl rollout status "deployment/${DEPLOYMENT_NAME}" -n "${NAMESPACE}" || true
echo ""

echo -e "${YELLOW}[2/3] rollout 历史版本:${NC}"
kubectl rollout history "deployment/${DEPLOYMENT_NAME}" -n "${NAMESPACE}"
echo ""

echo -e "${YELLOW}[3/3] 执行回滚 (kubectl rollout undo)...${NC}"
kubectl rollout undo "deployment/${DEPLOYMENT_NAME}" -n "${NAMESPACE}"

echo ""
echo -e "${GREEN}✅ 已触发回滚，等待 rollout 完成...${NC}"
kubectl rollout status "deployment/${DEPLOYMENT_NAME}" -n "${NAMESPACE}" --timeout=5m

echo ""
echo -e "${GREEN}=== 回滚完成 ===${NC}"
echo -e "回滚后历史版本："
kubectl rollout history "deployment/${DEPLOYMENT_NAME}" -n "${NAMESPACE}"
