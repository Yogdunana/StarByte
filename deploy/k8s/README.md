# StarByte Kubernetes 清单（可选）

**学校当前生产部署以 Docker Compose 为准**，见 `deploy/docker-compose.yml` 与 `docs/deployment.md`。
本目录是学习 / 预研模板，合入仓库**不会**让 `main` 自动往集群部署。

## 目录

```
deploy/k8s/
├── kustomization.yaml           # 默认清单（不含 Secret / Ingress）
├── namespace.yaml
├── configmap.yaml
├── secret.yaml.example          # 复制为 secret.yaml 后填写，勿提交
├── postgres-statefulset.yaml
├── redis-statefulset.yaml
├── minio-deployment.yaml
├── backend-deployment.yaml
├── backend-service.yaml
├── frontend-deployment.yaml
├── frontend-service.yaml
├── ingress.yaml                 # 可选，默认不 apply
├── rollback.sh                  # ./rollback.sh backend|frontend|all
├── blue-green/
│   ├── README.md
│   ├── *-deployment-{blue,green}.yaml
│   ├── service-blue-green.yaml
│   └── switch.sh
└── canary/
    ├── README.md
    ├── backend-canary-deployment.yaml
    ├── frontend-canary-deployment.yaml
    ├── canary-services.yaml
    └── canary-ingress.yaml
```

Helm Chart：`deploy/helm/starbyte/`。

## 密钥 key（Helm 与 raw k8s 相同）

| Secret key | 用途 |
|------------|------|
| `DB_PASSWORD` | Postgres `POSTGRES_PASSWORD` + 后端 `DB_PASSWORD` |
| `REDIS_PASSWORD` | Redis / 后端 |
| `MINIO_ACCESS_KEY` | MinIO `MINIO_ROOT_USER` + 后端 |
| `MINIO_SECRET_KEY` | MinIO `MINIO_ROOT_PASSWORD` + 后端 |
| `JWT_SECRET` | 后端 JWT |

镜像名与 Chart 一致：`ghcr.io/yogdunana/starbyte-backend`、`ghcr.io/yogdunana/starbyte-frontend`。

## 快速部署（实验集群）

```bash
cp deploy/k8s/secret.yaml.example deploy/k8s/secret.yaml
# 编辑 secret.yaml，填入真实 base64 值
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -k deploy/k8s/
kubectl get pods -n starbyte
```

Helm（生产覆盖禁止自建 Secret）：

```bash
# 先创建 Secret，再
helm upgrade --install starbyte deploy/helm/starbyte \
  --namespace starbyte --create-namespace \
  --values deploy/helm/starbyte/values-prod.yaml \
  --set backend.image.repository=ghcr.io/yogdunana/starbyte-backend \
  --set backend.image.tag=1.0.0 \
  --set frontend.image.repository=ghcr.io/yogdunana/starbyte-frontend \
  --set frontend.image.tag=1.0.0
```

Ingress 默认关闭。需要时再 `kubectl apply -f deploy/k8s/ingress.yaml`，不必上 Let's Encrypt。

## 回滚 / 蓝绿 / 金丝雀

```bash
./deploy/k8s/rollback.sh backend
./deploy/k8s/rollback.sh frontend
./deploy/k8s/rollback.sh all

./deploy/k8s/blue-green/switch.sh green
./deploy/k8s/blue-green/switch.sh status
```

金丝雀文档与文件名以 `canary/README.md` 为准（Service 在 `canary-services.yaml`）。
