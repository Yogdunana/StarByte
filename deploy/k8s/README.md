# StarByte Kubernetes 部署清单 - Issue #102

## 目录结构

```
deploy/k8s/
├── kustomization.yaml           # Kustomize 入口（不含 Secret）
├── namespace.yaml               # Namespace 定义
├── configmap.yaml               # 非敏感配置
├── secret.yaml.example          # Secret 模板（生产用 Sealed Secrets）
├── postgres-statefulset.yaml    # PostgreSQL StatefulSet + Service
├── redis-statefulset.yaml       # Redis StatefulSet + Service
├── minio-deployment.yaml        # MinIO Deployment + Service
├── backend-deployment.yaml      # 后端 Deployment（滚动更新）
├── backend-service.yaml         # 后端 Service
├── frontend-deployment.yaml     # 前端 Deployment（滚动更新）
├── frontend-service.yaml        # 前端 Service
├── ingress.yaml                 # Ingress（TLS + 路由）
├── rollback.sh                  # 回滚脚本
├── blue-green/                  # 蓝绿部署配置
│   ├── README.md
│   ├── backend-deployment-blue.yaml
│   ├── backend-deployment-green.yaml
│   ├── frontend-deployment-blue.yaml
│   ├── frontend-deployment-green.yaml
│   ├── service-blue-green.yaml
│   └── switch.sh
└── canary/                      # 金丝雀发布配置
    ├── README.md
    ├── backend-canary-deployment.yaml
    ├── frontend-canary-deployment.yaml
    ├── canary-services.yaml
    └── canary-ingress.yaml
```

## 快速部署

### 方式一：Kustomize（推荐）
```bash
# 1. 创建 Secret（生产环境用 Sealed Secrets 或 External Secrets）
cp deploy/k8s/secret.yaml.example deploy/k8s/secret.yaml
# 编辑 secret.yaml，填入真实密钥值
kubectl apply -f deploy/k8s/secret.yaml -n starbyte

# 2. 部署
kubectl apply -k deploy/k8s/

# 3. 查看状态
kubectl get pods -n starbyte
```

### 方式二：Helm（生产推荐）
```bash
helm install starbyte deploy/helm/starbyte \
  --namespace starbyte --create-namespace \
  --values deploy/helm/starbyte/values-prod.yaml
```

## 密钥管理

**禁止将真实密钥提交到 Git 仓库。** 生产环境请使用以下方案之一：

### Sealed Secrets（推荐）
```bash
# 安装 Sealed Secrets Controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.27.3/controller.yaml

# 创建加密的 Secret
echo -n 'starbyte123' | base64 | kubectl create secret generic starbyte-secret \
  --dry-run=client --from-literal=POSTGRES_PASSWORD=starbyte123 -o yaml | \
  kubeseal --format yaml > deploy/k8s/sealed-secret.yaml
```

### External Secrets（适合多集群）
```bash
# 从 AWS Secrets Manager / Vault 等外部密钥管理系统拉取
kubectl apply -f https://raw.githubusercontent.com/external-secrets/external-secrets/main/deploy/bundle.yaml
```

## 回滚
```bash
# 回滚后端到上一版本
./deploy/k8s/rollback.sh backend

# 回滚前端
./deploy/k8s/rollback.sh frontend

# 回滚所有
./deploy/k8s/rollback.sh all
```

## 部署策略对照

| 策略 | 适用场景 | 路径 |
|------|----------|------|
| 滚动更新 | 常规发布（默认） | `deploy/k8s/*.yaml` |
| 蓝绿部署 | 零停机发布 + 秒级回滚 | `deploy/k8s/blue-green/` |
| 金丝雀发布 | 渐进式发布 + 流量切分 | `deploy/k8s/canary/` |
| Helm | 多环境差异化部署 | `deploy/helm/starbyte/` |
