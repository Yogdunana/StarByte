# StarByte 金丝雀发布 - Issue #102

## 策略说明

金丝雀发布（Canary Release）逐步将新版本暴露给部分用户，先验证再全量推广。
本方案基于 Nginx Ingress Controller 的 `nginx.ingress.kubernetes.io/canary-*` 注解实现流量切分，无需额外组件。

## 使用方法

### 1. 初始状态
主 Ingress（`deploy/k8s/ingress.yaml`）指向稳定的 backend/frontend Deployment。

### 2. 部署金丝雀版本
```bash
kubectl apply -f deploy/k8s/canary/backend-canary-deployment.yaml
kubectl apply -f deploy/k8s/canary/frontend-canary-deployment.yaml
kubectl apply -f deploy/k8s/canary/backend-canary-service.yaml
kubectl apply -f deploy/k8s/canary/frontend-canary-service.yaml
```

### 3. 创建金丝雀 Ingress（初始 10% 流量）
```bash
kubectl apply -f deploy/k8s/canary/canary-ingress.yaml
```

### 4. 逐步增加流量
```bash
# 10% → 30%
kubectl annotate ingress starbyte-canary-ingress \
  nginx.ingress.kubernetes.io/canary-weight=30 \
  -n starbyte --overwrite

# 30% → 50%
kubectl annotate ingress starbyte-canary-ingress \
  nginx.ingress.kubernetes.io/canary-weight=50 \
  -n starbyte --overwrite

# 50% → 100%（全量）
kubectl annotate ingress starbyte-canary-ingress \
  nginx.ingress.kubernetes.io/canary-weight=100 \
  -n starbyte --overwrite
```

### 5. 完成发布
金丝雀 100% 验证通过后：
1. 更新主 Deployment 的 image tag 为金丝雀版本
2. 删除金丝雀资源：`kubectl delete -f deploy/k8s/canary/`

### 6. 回滚
```bash
# 将金丝雀流量降回 0
kubectl annotate ingress starbyte-canary-ingress \
  nginx.ingress.kubernetes.io/canary-weight=0 \
  -n starbyte --overwrite
# 删除金丝雀资源
kubectl delete -f deploy/k8s/canary/
```

## 按用户灰度（可选）
不使用百分比，而用 Header/Cookie 路由：
```yaml
annotations:
  nginx.ingress.kubernetes.io/canary-by-header: "X-Canary"
  nginx.ingress.kubernetes.io/canary-by-header-value: "true"
```
带 `X-Canary: true` 的请求走金丝雀版本。
