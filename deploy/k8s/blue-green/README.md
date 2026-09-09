# StarByte 蓝绿部署 - Issue #102

## 策略说明

蓝绿部署维护两个完全相同的环境（Blue / Green），同一时刻只有其中一个接收流量。
发布时将新版本部署到空闲环境，验证通过后切换流量，如有问题可秒级回滚。

## 使用方法

### 1. 初始部署（Blue 激活）
```bash
kubectl apply -f deploy/k8s/blue-green/backend-deployment-blue.yaml
kubectl apply -f deploy/k8s/blue-green/backend-deployment-green.yaml   # 副本数 0
kubectl apply -f deploy/k8s/blue-green/frontend-deployment-blue.yaml
kubectl apply -f deploy/k8s/blue-green/frontend-deployment-green.yaml  # 副本数 0
kubectl apply -f deploy/k8s/blue-green/service-blue-green.yaml          # 默认指向 Blue
```

### 2. 发布新版本（部署到 Green）
```bash
# 更新 Green Deployment 的 image tag
kubectl set image deployment/starbyte-backend-green \
  backend=ghcr.io/yogdunana/starbyte-backend:v2.0.0 -n starbyte
kubectl set image deployment/starbyte-frontend-green \
  frontend=ghcr.io/yogdunana/starbyte-frontend:v2.0.0 -n starbyte

# 等待 Green 就绪
kubectl rollout status deployment/starbyte-backend-green -n starbyte
kubectl rollout status deployment/starbyte-frontend-green -n starbyte
```

### 3. 切换流量到 Green
```bash
kubectl patch service starbyte-backend-bg -n starbyte \
  -p '{"spec":{"selector":{"version":"green"}}}'
kubectl patch service starbyte-frontend-bg -n starbyte \
  -p '{"spec":{"selector":{"version":"green"}}}'
```

### 4. 回滚到 Blue
```bash
kubectl patch service starbyte-backend-bg -n starbyte \
  -p '{"spec":{"selector":{"version":"blue"}}}'
kubectl patch service starbyte-frontend-bg -n starbyte \
  -p '{"spec":{"selector":{"version":"blue"}}}'
```

### 5. 使用脚本（推荐）
```bash
# 切换到 Green
./deploy/k8s/blue-green/switch.sh green

# 切换到 Blue（回滚）
./deploy/k8s/blue-green/switch.sh blue

# 查看当前状态
./deploy/k8s/blue-green/switch.sh status
```
