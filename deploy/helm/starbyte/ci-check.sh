#!/usr/bin/env bash
# helm lint + template 静态校验（CI 与本地共用）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
CHART="${ROOT}/deploy/helm/starbyte"
OUT="$(mktemp -d)"
trap 'rm -rf "$OUT"' EXIT

helm lint "$CHART"
helm lint "$CHART" -f "${CHART}/values-dev.yaml"
helm lint "$CHART" -f "${CHART}/values-prod.yaml"

echo "==> template values-dev"
helm template starbyte "$CHART" -n starbyte-dev -f "${CHART}/values-dev.yaml" > "${OUT}/dev.yaml"

echo "==> template values-prod"
helm template starbyte "$CHART" -n starbyte -f "${CHART}/values-prod.yaml" > "${OUT}/prod.yaml"

# 生产不得渲染 Secret
if grep -q '^kind: Secret$' "${OUT}/prod.yaml"; then
  echo "values-prod 不应渲染 Secret（secrets.create 必须为 false）" >&2
  exit 1
fi

# 默认不得渲染 Ingress
if grep -q '^kind: Ingress$' "${OUT}/prod.yaml"; then
  echo "values-prod 默认不应渲染 Ingress" >&2
  exit 1
fi
if grep -q '^kind: Ingress$' "${OUT}/dev.yaml"; then
  echo "values-dev 默认不应渲染 Ingress" >&2
  exit 1
fi

# 不得出现 Let's Encrypt / cert-manager 默认依赖
if grep -Eiq 'letsencrypt|cert-manager.io' "${OUT}/prod.yaml" "${OUT}/dev.yaml"; then
  echo "渲染结果不应默认依赖 Let's Encrypt / cert-manager" >&2
  exit 1
fi

# 镜像必须是 repository:tag，且 tag 不是完整 URL
if ! grep -q 'image: "ghcr.io/yogdunana/starbyte-backend:1.0.0"' "${OUT}/prod.yaml"; then
  echo "backend 镜像未按 repository+tag 拼接" >&2
  grep -n 'starbyte-backend' "${OUT}/prod.yaml" || true
  exit 1
fi
if ! grep -q 'image: "ghcr.io/yogdunana/starbyte-frontend:1.0.0"' "${OUT}/prod.yaml"; then
  echo "frontend 镜像未按 repository+tag 拼接" >&2
  exit 1
fi

# Secret key 对齐：开发 Secret 必须含后端实际读取的 key
for key in DB_PASSWORD REDIS_PASSWORD MINIO_ACCESS_KEY MINIO_SECRET_KEY JWT_SECRET; do
  if ! grep -q "^  ${key}:" "${OUT}/dev.yaml"; then
    echo "dev Secret 缺少 key ${key}" >&2
    exit 1
  fi
done

# 传入完整镜像 URL 作为 tag 必须失败
if helm template starbyte "$CHART" -n starbyte \
  --set backend.image.tag='ghcr.io/yogdunana/starbyte-backend:latest' \
  >"${OUT}/bad.yaml" 2>"${OUT}/bad.err"; then
  echo "应拒绝把完整镜像 URL 当作 image.tag" >&2
  exit 1
fi
if ! grep -q 'image.tag 只能是 tag' "${OUT}/bad.err"; then
  echo "tag 校验错误信息不匹配:" >&2
  cat "${OUT}/bad.err" >&2
  exit 1
fi

# 生产若误开 create 且无 allowInsecure，必须失败
if helm template starbyte "$CHART" -n starbyte -f "${CHART}/values-prod.yaml" \
  --set secrets.create=true \
  >"${OUT}/prod-secret.yaml" 2>"${OUT}/prod-secret.err"; then
  echo "prod + secrets.create=true 应失败" >&2
  exit 1
fi

# kustomize 默认清单可构建，且不含 Ingress/Secret
if command -v kustomize >/dev/null 2>&1; then
  echo "==> kustomize build"
  kustomize build "${ROOT}/deploy/k8s" > "${OUT}/kust.yaml"
elif command -v kubectl >/dev/null 2>&1; then
  echo "==> kubectl kustomize"
  kubectl kustomize "${ROOT}/deploy/k8s" > "${OUT}/kust.yaml"
fi
if [ -f "${OUT}/kust.yaml" ]; then
  if grep -q '^kind: Ingress$' "${OUT}/kust.yaml"; then
    echo "默认 kustomization 不应包含 Ingress" >&2
    exit 1
  fi
  if grep -q '^kind: Secret$' "${OUT}/kust.yaml"; then
    echo "默认 kustomization 不应包含 Secret" >&2
    exit 1
  fi
fi

echo "✅ helm lint / template / 安全断言通过"
