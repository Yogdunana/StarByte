# StarByte 蓝绿部署（学习模板）

与 Docker Compose 生产路径并行存在，**不要**在未规划的集群上当默认发布方式。

## 策略

维护 Blue / Green 两套 Deployment，Service 用 `selector.version` 切流量。

`switch.sh` 使用 JSON patch **只改** `/spec/selector/version`，并校验 `app` / `component` 仍在。
不要用只含 `version` 的 merge/strategic patch，那会整表替换 selector。

## 使用

### 1. 初始（Blue 接流量）

```bash
kubectl apply -f deploy/k8s/blue-green/backend-deployment-blue.yaml
kubectl apply -f deploy/k8s/blue-green/backend-deployment-green.yaml
kubectl apply -f deploy/k8s/blue-green/frontend-deployment-blue.yaml
kubectl apply -f deploy/k8s/blue-green/frontend-deployment-green.yaml
kubectl apply -f deploy/k8s/blue-green/service-blue-green.yaml
```

### 2. 发布到空闲色

```bash
kubectl set image deployment/starbyte-backend-green \
  backend=ghcr.io/yogdunana/starbyte-backend:v2.0.0 -n starbyte
kubectl set image deployment/starbyte-frontend-green \
  frontend=ghcr.io/yogdunana/starbyte-frontend:v2.0.0 -n starbyte
kubectl rollout status deployment/starbyte-backend-green -n starbyte
kubectl rollout status deployment/starbyte-frontend-green -n starbyte
```

### 3. 切换 / 回滚

```bash
./deploy/k8s/blue-green/switch.sh green
./deploy/k8s/blue-green/switch.sh status
# 回滚
./deploy/k8s/blue-green/switch.sh blue
```

滚动更新回滚（非蓝绿）见 `../rollback.sh`：

```bash
./deploy/k8s/rollback.sh backend
./deploy/k8s/rollback.sh frontend
./deploy/k8s/rollback.sh all
```
