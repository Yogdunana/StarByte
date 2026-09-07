# 部署文档

生产环境以 `deploy/docker-compose.yml` 为准。开发环境用 `deploy/docker-compose.dev.yml`（会把 Postgres/Redis/MinIO 端口打到主机）。

## Docker Compose（推荐）

```bash
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte
cp deploy/.env.example deploy/.env   # 按需改密码、JWT_SECRET
docker compose -f deploy/docker-compose.yml up -d --build
docker compose -f deploy/docker-compose.yml ps
```

启动后：

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost/ （容器 80） |
| API | http://localhost:8080/api/v1 |
| 健康检查 | http://localhost:8080/health 、`/health/ready` |
| Metrics | http://localhost:8080/metrics |
| Swagger（非生产） | http://localhost:8080/swagger/index.html |
| MinIO API | http://localhost:9000 |

首次启动后端会跑迁移。种子数据（角色/权限/演示账号）需在能连上库后执行：

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

## Nginx 反向代理（示例）

```nginx
server {
    listen 80;
    server_name starbyte.work;
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

用 Let's Encrypt 或学校证书终止 TLS，再反代到上面的 80/8080。不要单独做 `auth.` 子域，认证回调走 `https://starbyte.work/api/v1/auth/...`。

## 数据库备份与恢复

```bash
# 备份
docker exec starbyte-postgres pg_dump -U starbyte starbyte > starbyte-$(date +%F).sql

# 恢复
cat starbyte-2026-09-07.sql | docker exec -i starbyte-postgres psql -U starbyte starbyte
```

MinIO 桶与 Postgres 一起备份。恢复后执行 `make migrate-up` 确认 schema 版本。
