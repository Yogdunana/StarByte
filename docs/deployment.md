# 部署文档

生产环境以 `deploy/docker-compose.yml` 为准。开发环境用 `deploy/docker-compose.dev.yml`（会把 Postgres/Redis/MinIO 端口打到主机）。

应急运维：`deploy/cli/starbyte`（安装后可用 `starbyte` / `sb`，类似宝塔 `bt`）。

## 新服务器清单（校园网 / 内网裸机）

目标：格式化新机 → 只走国内可达源拉依赖 → 恢复导出数据 → `up`。**不必再对 Dockerfile 做 ad-hoc sed。**

1. **安装 Docker Engine + Compose 插件**（发行版官方源或学校镜像）。确认 `docker compose version` 可用。
2. **（可选）配置 registry-mirrors**，见下方「daemon.json」。先 `docker pull hello-world` 自测，**不要**把未验证的镜像站写进 compose。
3. **取得代码**：`git clone`，或把发布 zip 解到目标目录。若用 zip 覆盖已有目录：`bash deploy/update-from-zip.sh /path/to/StarByte.zip`（会保留 `deploy/.env`，且**绝不**触碰 named volumes）。
4. **复制环境变量**：`cp deploy/.env.example deploy/.env`，改密码与 `JWT_SECRET`。内网保持 `SKIP_BUCKET_CREATE=true`。`CAS_AUTHSERVER_IP` 默认 `10.100.14.250`（写入 backend `extra_hosts`）。
5. **恢复数据（若有导出）**：先 `up` 出 postgres/minio，再导入 SQL / 拷回 MinIO 数据，见「数据库备份与恢复」。
6. **构建并启动**：`docker compose -f deploy/docker-compose.yml up -d --build`
7. **首次建桶**：`SKIP_BUCKET_CREATE=true` 时后端不会自动建桶。在 MinIO 控制台（或 `mc`）**手动创建一次** `starbyte` 桶。
8. **健康检查**：`curl -sf http://127.0.0.1:8080/health`；浏览器打开 `http://<服务器IP>/`。或 `starbyte status` / `starbyte doctor`。
9. **更新应用时保留 volumes**：只重建前后端，例如 `starbyte rebuild all`（内部 `up --no-deps --force-recreate`），或  
   `docker compose -f deploy/docker-compose.yml up -d --no-deps --force-recreate backend frontend`。  
   **不要** `down -v`，不要删除 `postgres-data` / `redis-data` / `minio-data`。

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

前端与基础设施仍用 docker.io 官方 tag，仓库**不写死**未验证的国内 registry / 华为 SWR：

| 用途 | 镜像 |
|------|------|
| 前端构建 | `node:20-alpine` |
| 前端生产 | `nginx:1.27-alpine` |
| 数据库 | `postgres:16-alpine` |
| 缓存 | `redis:7-alpine` |
| 对象存储 | `${MINIO_IMAGE:-minio/minio:latest}` |

若 `docker pull` 失败：用已配置的 Docker Hub 镜像源拉取，或在可联网机器上 `docker pull` 后 `tag` / `save` 导入内网。

### MinIO 镜像与本地复用

生产机上曾出现：compose 钉死 `minio/minio:RELEASE.2024-08-17T01-16-21Z`，经校园镜像站拉取得到 **403**，而本机已有健康的 `minio/minio:latest` 却用不上。

当前 compose：

- 默认 `minio/minio:latest`，可用环境变量 `MINIO_IMAGE=...` 覆盖（写入 `deploy/.env`）
- `pull_policy: if_not_present`：本地已有该 tag 时**不强制拉取**
- postgres / redis 同样 `if_not_present`，避免镜像站抖动时把 `up` 卡死

CI 只 `docker build` 前后端生产 target，不 `compose pull` MinIO，不受此默认 tag 影响。

若必须钉 RELEASE：在 `.env` 设 `MINIO_IMAGE=minio/minio:RELEASE.2024-08-17T01-16-21Z`，并保证该 tag 已在本机或可拉取。

### daemon.json 镜像源（可选，先自测）

**不要**把某一个未验证的镜像站写进 `docker-compose.yml` 的 `image:`。校园常见镜像站会对部分仓库返回 403（例如某机上的 `docker.xuanyuan.me` 拉 `minio/minio:RELEASE.*`）。

若学校或机房提供可用加速，在宿主机配置 Docker，而不是改仓库：

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json >/dev/null <<'EOF'
{
  "registry-mirrors": [
    "https://example-campus-mirror.invalid"
  ]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker
```

把 `example-campus-mirror.invalid` 换成**本机已验证**的地址。可选来源（均须自行 `docker pull` 验证，失效很常见）：

- 学校 / 机房自建 registry-mirrors
- 云厂商容器镜像加速（须登录控制台获取专属地址）
- 社区公开加速（地址经常更换或 403，**禁止**当仓库默认）

验证：

```bash
docker info | grep -A20 'Registry Mirrors'
docker pull hello-world
```

拉失败时改用离线导入：在可访问 Docker Hub 的机器 `docker pull` + `docker save`，拷到内网 `docker load`。postgres / redis / node / nginx / minio 均可如此导入后再 `compose up --build`（`if_not_present` 会复用本地镜像）。

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

## 应急 CLI（`starbyte` / `sb`）

SSH 登录后的轻量运维入口，依赖 **bash + docker compose**，不读、不打印密钥值。

```bash
# 在仓库根执行（一次性安装到 PATH，并写入 /etc/starbyte.conf）
sudo bash deploy/cli/starbyte install

# 或手动：
sudo install -m 0755 deploy/cli/starbyte /usr/local/bin/starbyte
sudo ln -sfn /usr/local/bin/starbyte /usr/local/bin/sb
echo "STARBYTE_HOME=$(pwd)" | sudo tee /etc/starbyte.conf
```

之后任意目录：

| 命令 | 作用 |
|------|------|
| `starbyte status` / `sb status` | `compose ps` + `:8080/health`、`:80` 短检查 |
| `starbyte logs [服务] [--tail N]` | 看日志 |
| `starbyte restart [服务\|all]` | 重启容器（不重建镜像） |
| `starbyte rebuild [backend\|frontend\|all]` | `build` + `up --no-deps --force-recreate`，避免顺带强拉 MinIO/Postgres |
| `starbyte env-check` | 只检查 `deploy/.env` **键名**是否存在且非空，不打印值 |
| `starbyte backup create\|list\|preview\|restore\|drill` | 经 backend 容器走托管备份（gzip / AES / 预览 / 失败告警） |
| `starbyte backup emergency` | 主机明文 `pg_dump` → `/var/backups/starbyte/`（backend 不可用时的兜底） |
| `starbyte doctor` | 校园部署常见问题：镜像站 403、`SKIP_BUCKET_CREATE`、CAS 回调 / authserver 内网解析、80 端口 |

`rebuild` 对应生产上已验证的绕过方式：镜像站 403 时不要对 MinIO 做 `compose up` 全量拉取。

## 用 zip 更新代码（保留 .env 与 volumes）

```bash
bash deploy/update-from-zip.sh /path/to/StarByte-main.zip
starbyte rebuild all
```

脚本会把 zip 解到仓库上，**始终保留现有 `deploy/.env`**，且不执行任何 `docker volume` / `down -v`。`deploy/docker-compose.yml` **会被 zip 覆盖**（不要靠机器上手改 compose）。CAS 所需的 `extra_hosts` 已写进仓库 compose，zip 更新后 `rebuild` 即可保持 `authserver.smbu.edu.cn` → `CAS_AUTHSERVER_IP`。

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
- **容器解析**：backend 带 `extra_hosts`，`authserver.smbu.edu.cn` → `${CAS_AUTHSERVER_IP:-10.100.14.250}`。校园 DNS 常返回不可达 IPv6，callback 会 wget 超时约 20s（nginx 502）。改 IP 后必须 `starbyte rebuild backend`（或 `--force-recreate`）才会写入容器 `/etc/hosts`
- 环境变量见 `deploy/.env.example` / `backend/.env.example` 的 `CAS_*`
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
# 推荐：与应用同一套 custom dump + gzip + 可选 AES
starbyte backup create
starbyte backup preview <id>
# 演练到独立库（先 createdb starbyte_drill；不是 PITR）
starbyte backup drill <id> --dbname starbyte_drill --confirm DRILL

# backend 挂了才用明文应急
starbyte backup emergency

# 覆盖当前应用库（危险）
starbyte backup restore <id> --confirm RESTORE
```

WAL / PITR 不在应用内：需要数据库主机配置 `wal_level`、归档命令和 `pg_basebackup`。

MinIO 桶与 Postgres 一起备份。Named volumes 典型路径可用 `docker volume inspect` 查看（compose 项目名多为 `deploy`，卷名类似 `deploy_postgres-data`）。恢复后确认迁移版本；`SKIP_BUCKET_CREATE=true` 时桶需已存在。
