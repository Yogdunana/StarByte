#!/bin/bash
# StarByte 蓝绿部署流量切换脚本 - Issue #102
# 用法:
#   ./switch.sh blue    切换流量到 Blue
#   ./switch.sh green   切换流量到 Green
#   ./switch.sh status  查看当前状态
set -euo pipefail

NAMESPACE="${STARBYTE_NAMESPACE:-starbyte}"

case "${1:-status}" in
  blue|green)
    TARGET="$1"
    echo "🔄 切换流量到 ${TARGET}..."

    # 扩容目标环境
    echo "  → 扩容 ${TARGET} 环境..."
    kubectl scale deployment "starbyte-backend-${TARGET}" --replicas=2 -n "$NAMESPACE"
    kubectl scale deployment "starbyte-frontend-${TARGET}" --replicas=2 -n "$NAMESPACE"

    # 等待就绪
    echo "  → 等待 ${TARGET} 就绪..."
    kubectl rollout status "deployment/starbyte-backend-${TARGET}" -n "$NAMESPACE" --timeout=3m
    kubectl rollout status "deployment/starbyte-frontend-${TARGET}" -n "$NAMESPACE" --timeout=3m

    # 切换 Service selector
    echo "  → 切换 Service selector → version=${TARGET}..."
    kubectl patch service starbyte-backend-bg -n "$NAMESPACE" \
      -p "{\"spec\":{\"selector\":{\"version\":\"${TARGET}\"}}}"
    kubectl patch service starbyte-frontend-bg -n "$NAMESPACE" \
      -p "{\"spec\":{\"selector\":{\"version\":\"${TARGET}\"}}}"

    # 缩容另一侧
    OTHER=$([ "$TARGET" = "blue" ] && echo "green" || echo "blue")
    echo "  → 缩容 ${OTHER} 环境（保留 0 副本）..."
    kubectl scale deployment "starbyte-backend-${OTHER}" --replicas=0 -n "$NAMESPACE"
    kubectl scale deployment "starbyte-frontend-${OTHER}" --replicas=0 -n "$NAMESPACE"

    echo "✅ 流量已切换到 ${TARGET}"
    ;;

  status)
    echo "📊 蓝绿部署状态 (namespace: ${NAMESPACE})"
    echo "──────────────────────────────────────────"
    for svc in starbyte-backend-bg starbyte-frontend-bg; do
      SELECTOR=$(kubectl get svc "$svc" -n "$NAMESPACE" \
        -o jsonpath='{.spec.selector.version}' 2>/dev/null || echo "N/A")
      echo "  ${svc} → 指向: ${SELECTOR:-unknown}"
    done
    echo ""
    echo "Deployments:"
    kubectl get deployments -n "$NAMESPACE" \
      -l 'app in (starbyte),component in (backend,frontend)' \
      -o custom-columns=NAME:.metadata.name,REPLICAS:.spec.replicas,READY:.status.readyReplicas,VERSION:.metadata.labels.version \
      2>/dev/null || echo "  (无部署)"
    ;;

  *)
    echo "用法: $0 {blue|green|status}"
    exit 1
    ;;
esac
