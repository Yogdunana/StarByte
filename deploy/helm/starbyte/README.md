# StarByte Helm Chart

学校默认部署仍是 `deploy/docker-compose.yml`。本 Chart 供实验集群使用，**不会**随 `main` 自动发布。

```bash
# 开发（允许 chart 自建本地 Secret）
helm upgrade --install starbyte . \
  -n starbyte-dev --create-namespace \
  -f values-dev.yaml

# 生产：先创建 Secret，再
helm upgrade --install starbyte . \
  -n starbyte --create-namespace \
  -f values-prod.yaml \
  --set backend.image.repository=ghcr.io/yogdunana/starbyte-backend \
  --set backend.image.tag=sha-abc1234 \
  --set frontend.image.repository=ghcr.io/yogdunana/starbyte-frontend \
  --set frontend.image.tag=sha-abc1234
```

`image.tag` 只能是 tag。传入完整镜像 URL 会被 helper 拒绝。

静态检查：`./ci-check.sh`（需要本机 `helm`）。
