# 部署文档

生产环境以 `deploy/docker-compose.yml` 为准。开发环境用 `deploy/docker-compose.dev.yml`（会把 Postgres/Redis/MinIO 端口打到主机）。

## Docker Compose（推荐）

```bash
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte
cp deploy/.env.example deploy/.env   # 按需改密码、JWT_SECRET；内网保持 SKIP_BUCKET_CREATE=true
docker compose -f deploy/docker-compose.yml up -d --build
docker compose -f deploy/docker-compose.yml ps
```

### 校园网 / 内网构建

校园服务器通常无法直连 GitHub、`dl.min.io`、`proxy.golang.org`。后端 `Dockerfile` 已按校园默认写好，**不必再 sed**：

- **golang-migrate v4.17.0**：经 `https://ghfast.top/https://github.com/...` 下载
- **Go modules**：`go env -w GOPROXY=https://goproxy.cn,direct` 后再 `go mod download`
- **MinIO mc**：镜像内是 `exit 0` 占位脚本，不从 `dl.min.io` 拉客户端。生产请设 `SKIP_BUCKET_CREATE=true`（`deploy/.env.example` 默认已是），在 MinIO 控制台手动创建一次 `starbyte` 桶

前端基础镜像仍是 docker.io 的 `node:20-alpine` / `nginx:1.27-alpine`。仓库不写死未验证的国内 registry。若 `docker pull` 失败，用已配置的 Docker Hub 镜像源拉取，或在可联网机器上 `docker pull` 后 `tag` / `save` 导入内网。

启动后：

| 服务 | 地址 |
|------|------|
| 前端 | 漏测：校园网 `http://<服务器IP>/`（容器 80）；有域名后再绑 `starbyte.smbu.edu.cn` |
| API | http://localhost:8080/api/v1 |
| 健康检查 | http://localhost:8080/health 、`/health/ready` |
| Metrics | http://localhost:8080/metrics |
| Swagger（非生产） | http://localhost:8080/swagger/index.html |
| MinIO API | http://localhost:9000 |

首次启动后端会跑迁移。生产 Postgres **不映射主机端口**，在宿主机执行 `APP_ENV=prod make seed` 会连不上库。请在能访问 `postgres` 服务的网络里跑种子（跳板机映射 5432，或一次性容器加入 compose 网络），并注入 `DB_HOST` / `DB_USER` / `DB_PASSWORD` / `JWT_SECRET` 等（见 `backend/.env.example`）。

手动部署（库端口对本机可见）时：

```bash
APP_ENV=prod make seed
```

默认账号：`admin/admin123`（社长，学号 `20210001`）、`test/test123`（会员，学号 `20210002`）。

## 手动部署

1. PostgreSQL 16、Redis 7、MinIO。
2. 后端：`cd backend && go build -o server ./cmd/server`，设置 `APP_ENV=prod`、`CONFIG_PATH` 与 `DB_*` / `REDIS_*` / `MINIO_*` / `JWT_SECRET`。
3. 迁移：`make migrate-up DATABASE_URL=postgres://...`。
4. 种子：`APP_ENV=prod make seed`。
5. 前端：`cd frontend && npm ci && npm run build`，用 Nginx 托管 `dist/` 并把 `/api/` 反代到后端。

## 学校统一认证（漏测先用 IP）

- 登录页「学校统一认证」跳到 `https://authserver.smbu.edu.cn/authserver/login`
- **漏测没有域名**：用校园网 IP 打开系统（`http://<服务器IP>/`）。`service` / 回跳按访问 Host 自动拼，信息化备案：
  `http://<服务器IP>/api/v1/auth/cas/callback`
- 有 `starbyte.smbu.edu.cn` 后再改备案，或设 `CAS_SERVICE_URL` / `CAS_FRONTEND_URL`
- 环境变量见 `backend/.env.example` 的 `CAS_*`
- 外网 `starbyte.com` 检测校内 IP 后 301 属二期

## Nginx 反向代理（示例）

```nginx
server {
    listen 80;
    server_name _;
    root /var/www/starbyte/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Request-ID $request_id;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## HTTPS

漏测可先 HTTP + IP。有证书后再终止 TLS。不要单独做 `auth.` 子域，CAS 回调与站点同 Host：`http(s)://<IP或域名>/api/v1/auth/cas/callback`。

## 数据库备份与恢复

```bash
# 备份
docker exec starbyte-postgres pg_dump -U starbyte starbyte > starbyte-$(date +%F).sql

# 恢复
cat starbyte-2026-09-07.sql | docker exec -i starbyte-postgres psql -U starbyte starbyte
```

MinIO 桶与 Postgres 一起备份。恢复后执行 `make migrate-up` 确认 schema 版本。
