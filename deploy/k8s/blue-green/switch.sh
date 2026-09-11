#!/usr/bin/env bash
# 蓝绿流量切换。
#   ./switch.sh green
#   ./switch.sh blue
#   ./switch.sh status
#
# #145 的 merge patch 只写 version，会整表替换 selector，丢掉 app/component。
# 这里用 JSON patch 只改 /spec/selector/version，并校验其余 label 仍在。
set -euo pipefail

NAMESPACE="${STARBYTE_NAMESPACE:-starbyte}"

assert_selector() {
  local svc="$1"
  local component="$2"
  local expected_version="$3"
  local app version comp
  app="$(kubectl get svc "$svc" -n "$NAMESPACE" -o jsonpath='{.spec.selector.app}')"
  comp="$(kubectl get svc "$svc" -n "$NAMESPACE" -o jsonpath='{.spec.selector.component}')"
  version="$(kubectl get svc "$svc" -n "$NAMESPACE" -o jsonpath='{.spec.selector.version}')"
  if [[ "$app" != "starbyte" || "$comp" != "$component" || "$version" != "$expected_version" ]]; then
    echo "❌ ${svc} selector 异常: app=${app} component=${comp} version=${version}" >&2
    echo "   期望: app=starbyte component=${component} version=${expected_version}" >&2
    exit 1
  fi
}

patch_version() {
  local svc="$1"
  local component="$2"
  local target="$3"
  # 只替换 version，避免 wipe app/component
  kubectl patch service "$svc" -n "$NAMESPACE" --type=json \
    -p "[{\"op\":\"replace\",\"path\":\"/spec/selector/version\",\"value\":\"${target}\"}]"
  assert_selector "$svc" "$component" "$target"
}

case "${1:-status}" in
  blue|green)
    TARGET="$1"
    echo "🔄 切换流量到 ${TARGET}..."

    echo "  → 扩容 ${TARGET} 环境..."
    kubectl scale deployment "starbyte-backend-${TARGET}" --replicas=2 -n "$NAMESPACE"
    kubectl scale deployment "starbyte-frontend-${TARGET}" --replicas=2 -n "$NAMESPACE"

    echo "  → 等待 ${TARGET} 就绪..."
    kubectl rollout status "deployment/starbyte-backend-${TARGET}" -n "$NAMESPACE" --timeout=3m
    kubectl rollout status "deployment/starbyte-frontend-${TARGET}" -n "$NAMESPACE" --timeout=3m

    echo "  → JSON patch selector.version=${TARGET}（保留 app/component）..."
    patch_version starbyte-backend-bg backend "$TARGET"
    patch_version starbyte-frontend-bg frontend "$TARGET"

    OTHER=$([ "$TARGET" = "blue" ] && echo "green" || echo "blue")
    echo "  → 缩容 ${OTHER} 环境（0 副本）..."
    kubectl scale deployment "starbyte-backend-${OTHER}" --replicas=0 -n "$NAMESPACE"
    kubectl scale deployment "starbyte-frontend-${OTHER}" --replicas=0 -n "$NAMESPACE"

    echo "✅ 流量已切换到 ${TARGET}，selector 仍含 app/component"
    ;;

  status)
    echo "📊 蓝绿部署状态 (namespace: ${NAMESPACE})"
    echo "──────────────────────────────────────────"
    for svc in starbyte-backend-bg starbyte-frontend-bg; do
      SELECTOR="$(kubectl get svc "$svc" -n "$NAMESPACE" \
        -o jsonpath='app={.spec.selector.app} component={.spec.selector.component} version={.spec.selector.version}' \
        2>/dev/null || echo "N/A")"
      echo "  ${svc} → ${SELECTOR}"
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
