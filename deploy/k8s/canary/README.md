# StarByte 金丝雀发布（学习模板）

依赖 **Nginx Ingress Controller** 的 `nginx.ingress.kubernetes.io/canary-*`。
学校内网默认没有 Ingress / Let's Encrypt，不要把本目录当默认部署路径。

## 文件名（与仓库一致）

| 用途 | 实际文件 |
|------|----------|
| 后端金丝雀 Deployment | `backend-canary-deployment.yaml` |
| 前端金丝雀 Deployment | `frontend-canary-deployment.yaml` |
| 前后端 Service（合并） | `canary-services.yaml` |
| 金丝雀 Ingress | `canary-ingress.yaml` |

没有 `backend-canary-service.yaml` / `frontend-canary-service.yaml`。

## 使用

主流量仍走 `deploy/k8s/` 的稳定 Deployment。需要七层入口时再单独 `kubectl apply -f deploy/k8s/ingress.yaml`。

```bash
kubectl apply -f deploy/k8s/canary/backend-canary-deployment.yaml
kubectl apply -f deploy/k8s/canary/frontend-canary-deployment.yaml
kubectl apply -f deploy/k8s/canary/canary-services.yaml
kubectl apply -f deploy/k8s/canary/canary-ingress.yaml
```

调权重：

```bash
kubectl annotate ingress starbyte-canary-ingress \
  nginx.ingress.kubernetes.io/canary-weight=30 \
  -n starbyte --overwrite
```

回滚：权重改回 `0`，或 `kubectl delete -f deploy/k8s/canary/`。

滚动更新回滚（稳定轨）使用：

```bash
./deploy/k8s/rollback.sh backend
./deploy/k8s/rollback.sh frontend
./deploy/k8s/rollback.sh all
```

按 Header 灰度（可选）：

```yaml
annotations:
  nginx.ingress.kubernetes.io/canary-by-header: "X-Canary"
  nginx.ingress.kubernetes.io/canary-by-header-value: "true"
```
