# 部署文档

生产环境以 `deploy/docker-compose.yml` 为准。开发环境用 `deploy/docker-compose.dev.yml`（会把 Postgres/Redis/MinIO 端口打到主机）。

应急运维：`deploy/cli/starbyte`（安装后可用 `starbyte` / `sb`，类似宝塔 `bt`）。

> **密钥策略（重要）**：仓库不再为任何口令提供默认值。`POSTGRES_PASSWORD` /
> `REDIS_PASSWORD` / `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` / `JWT_SECRET` /
> `STARBYTE_CONFIG_ENCRYPTION_KEY` 全部由 CLI 在安装时**自动随机生成**。
> 生成结果**只在安装终端打印一次**（不写日志、不落凭证文件）。**不要手填、也不要
> `cp .env.example .env` 后自己编**——compose 用 `${VAR:?}` 强制必填，没走生成流程会直接报错退出。
>
> 口令的唯一持久副本是 `deploy/.env`（生成时权限 0600；Windows 上权限位不生效，靠不要提交来保护）。
> `deploy/.env.credentials` 只留一行「去哪儿再读」的提示，**不含明文**，可以安全保留或删除。
> 忘了口令就跑 `starbyte credentials`（会在终端重打印，请在可信终端执行）。
>
> 日常运维命令（`status` / `logs` / `doctor` / `env-check` / `backup`）不打印任何密钥值；
> 只有 `starbyte init` 与 `starbyte credentials` 会显示。

## 新服务器清单（校园网 / 内网裸机）

目标：格式化新机 → 只走国内可达源拉依赖 → 恢复导出数据 → `up`。**不必再对 Dockerfile 做 ad-hoc sed。**

1. **安装 Docker Engine + Compose 插件**（发行版官方源或学校镜像）。确认 `docker compose version` 可用。
2. **（可选）配置 registry-mirrors**，见下方「daemon.json」。先 `docker pull hello-world` 自测，**不要**把未验证的镜像站写进 compose。
3. **取得代码**：`git clone`，或把发布 zip 解到目标目录。若用 zip 覆盖已有目录：`bash deploy/update-from-zip.sh /path/to/StarByte.zip`（会保留 `deploy/.env`，且**绝不**触碰 named volumes）。脚本已做供应链加固：**基础设施文件（Dockerfile、compose、CLI、entrypoint、configs、nginx 配置、`.github`、Makefile 等）永不覆盖**，并在发现 zip 含这些文件时告警；建议同时提供 `StarByte.zip.sha256` 做完整性校验（见下方「用 zip 更新代码」）。
4. **安装 CLI 并生成随机密钥**：`sudo bash deploy/cli/starbyte install`。
   首次安装时它会自动按 `deploy/.env.example` 生成 `deploy/.env`，把
   `POSTGRES_PASSWORD` / `REDIS_PASSWORD` / `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` /
   `JWT_SECRET` / `STARBYTE_CONFIG_ENCRYPTION_KEY` 全部填成随机值，并**在终端打印一次**。
   **立即抄录口令（建议直接进密码管理器）**——这是唯一一次明文展示，不会落进任何文件。
   忘了就再跑 `starbyte credentials`（在可信终端里）。
   之后按需编辑 `deploy/.env` 里**非密钥**项（CAS IP、`SKIP_BUCKET_CREATE`、`MINIO_IMAGE` 等）；
   `starbyte init --force` 会**替换全部密钥项**并保留非密钥键。注意：换掉库/缓存/对象存储的口令后，
   已初始化的数据卷仍用旧口令，需要同步改或清卷重来（会丢数据）——生产上更稳妥是重新走一次全新部署。
5. **恢复数据（若有导出）**：先 `up` 出 postgres/minio，再导入 SQL / 拷回 MinIO 数据，见「数据库备份与恢复」。
6. **构建并启动**：`docker compose -f deploy/docker-compose.yml up -d --build`
7. **首次建桶**：`SKIP_BUCKET_CREATE=true` 时后端不会自动建桶。用第 4 步生成的口令登录 MinIO 控制台（`ssh -L 9001:127.0.0.1:9001` 转发后访问 `http://localhost:9001`），**手动创建一次** `starbyte` 桶。
8. **健康检查**：`curl -sf http://127.0.0.1:8080/health`；浏览器打开 `http://<服务器IP>/`。或 `starbyte status` / `starbyte doctor`。
9. **更新应用时保留 volumes**：只重建前后端，例如 `starbyte rebuild all`（内部 `up --no-deps --force-recreate`），或  
   `docker compose -f deploy/docker-compose.yml up -d --no-deps --force-recreate backend frontend`。  
   **不要** `down -v`，不要删除 `postgres-data` / `redis-data` / `minio-data`。

> **上线前必须完成 HTTPS**（见「HTTPS」一节）。纯 HTTP 会在校园网内明文承载登录口令
> 与 Bearer token，同网段可 ARP 欺骗 / 流量镜像窃取；与固定的 JWT 密钥叠加构成直接沦陷路径。

## Docker Compose（推荐）

```bash
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte

# 1) 安装 CLI；首次会自动随机生成 deploy/.env 并打印口令（请立即抄录）
sudo bash deploy/cli/starbyte install

# 2) 校验密钥已就位（只报键名，不打印值）
starbyte env-check

# 3) 按需编辑非密钥项：CAS_AUTHSERVER_IP / MINIO_IMAGE / SKIP_BUCKET_CREATE ...
$EDITOR deploy/.env

# 4) 启动（务必带 --env-file，否则 compose 读不到 deploy/.env）
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --build
docker compose --env-file deploy/.env -f deploy/docker-compose.yml ps
```

> 手动部署（不走 CLI）时也**不要**手填密钥。直接 `up` 会看到
> `POSTGRES_PASSWORD 未设置，请先运行 starbyte init` 之类的报错，这是预期的失败关闭行为。

### 校园网 / 内网构建

校园服务器通常无法直连 GitHub、`dl.min.io`、`proxy.golang.org`。后端 `Dockerfile` 已按校园默认写好，**不必再 sed**：

- **golang-migrate v4.17.0**：在构建阶段经 `goproxy.cn` `go install`（不走 ghfast.top）
- **Go modules**：`go env -w GOPROXY=https://goproxy.cn,direct` 后再 `go mod download`
- **MinIO mc**：镜像内是 `exit 0` 占位脚本，不从 `dl.min.io` 拉客户端。生产请设 `SKIP_BUCKET_CREATE=true`（`deploy/.env.example` 默认已是），在 MinIO 控制台手动创建一次 `starbyte` 桶

前端与基础设施仍用 docker.io 官方 tag，仓库**不写死**未验证的国内 registry / 华为 SWR：

| 用途 | 镜像 |
|------|------|
| 后端构建 | `golang:1.25-alpine` |
| 后端运行 / mc 占位 | `alpine:3.21` |
| 前端构建 | `node:22-alpine` |
| 前端生产 | `nginx:1.27-alpine` |
| 数据库 | `postgres:16-alpine` |
| 缓存 | `redis:7-alpine` |
| 对象存储 | `${MINIO_IMAGE:-minio/minio:latest}` |

> **Go 工具链下限是 1.25**：`go.mod` 的 `go` 指令已提升到 `1.25.0`（`github.com/xuri/excelize/v2 v2.11.0`
> 与 `golang.org/x/net v0.55+` 的最低要求，二者分别修掉 CVE-2026-54063 / CVE-2026-59161 与
> GO-2026-5030 的解析型 DoS）。本地开发也需 Go ≥ 1.25。
>
> **镜像 tag 未在 CI 之外实测**：`alpine:3.21` / `golang:1.25-alpine` / `node:22-alpine` 按官方发布
> 节奏选定。首次构建前请先 `docker pull` 验证；拉不到就按下一节的办法离线导入。
> `minio/minio:latest` 仍不可复现 —— 要可复现请显式设 `MINIO_IMAGE=<具体 tag 或 digest>`。

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

启动后（生产 compose 已把后端 8080 与 MinIO 9000 **只绑回环**，校园网其它主机无法直连）：

| 服务 | 地址 |
|------|------|
| 前端 | 校园网 `http://<服务器IP>/`（宿主机 80 → 容器非 root nginx 的 8080） |
| API | 仅本机 `http://127.0.0.1:8080/api/v1`（容器间走 `http://backend:8080`） |
| 健康检查 | http://127.0.0.1:8080/health 、`/health/ready` |
| Metrics | 默认 **404**（未设 `METRICS_TOKEN` 时一律拒绝，只有显式 `APP_ENV=dev\|test` 的本地开发才放行）。需抓取时在 `.env` 设 `METRICS_TOKEN`，带 `Authorization: Bearer <token>` 访问 |
| Swagger（非生产） | http://127.0.0.1:8080/swagger/index.html（`APP_ENV=prod` 时关闭） |
| MinIO API | 仅本机 `http://127.0.0.1:9000`；控制台请用 `ssh -L 9001:127.0.0.1:9001` 转发后访问 |

首次启动后端会跑迁移。生产 Postgres **不映射主机端口**，在宿主机执行 `APP_ENV=prod make seed` 会连不上库。请在能访问 `postgres` 服务的网络里跑种子（跳板机映射 5432，或一次性容器加入 compose 网络），并注入 `DB_HOST` / `DB_USER` / `DB_PASSWORD` / `JWT_SECRET` 等（见 `backend/.env.example`）。

手动部署（库端口对本机可见）时：

```bash
APP_ENV=prod make seed
```

**种子账号（已加固）**：

- 生产**不再创建 `test` 账号**，也不再使用 `admin123`。
- 判定与后端配置一致（fail-closed）：**只有显式 `APP_ENV=dev|test` 才使用开发便利口令**
  （`admin/admin123`、`test/test123`）；`APP_ENV` 未设置一律按生产处理 —— 因为
  `config.Load` 在 `APP_ENV` 为空时也是按生产校验密钥的，两边不能给出相反结论。
  `make seed` 已默认带上 `APP_ENV=dev`，本地开发不受影响。
- `admin` 初始口令来源：优先读环境变量 `SEED_ADMIN_PASSWORD`；没有时由 `crypto/rand`
  现场生成 **24 位强随机口令**，并在执行 `make seed` 的终端**打印一次**：

  ```
  ============================================================
  [StarByte] 生产环境首次播种：admin 账号已创建
  用户名: admin
  初始密码: <24 位随机口令>
  请立即登录并修改密码；本密码只显示这一次。
  ============================================================
  ```

- 若 `admin` 已存在，只打印「已存在，跳过」且**不打印口令**（避免重复播种泄露）。
- **立即抄录并登录改密**。改密后服务端会吊销该账号全部在线会话与 refresh token。
- 非生产环境（`APP_ENV != prod`）保持 `admin/admin123`、`test/test123` 以方便开发，
  但会打印一行显式提示这是开发口令。

## 应急 CLI（`starbyte` / `sb`）

SSH 登录后的轻量运维入口，依赖 **bash + docker compose**。日常命令不读、不打印密钥值；
只有 `init` 与 `credentials` 会显示口令，因为它们是「安装时由系统自动随机生成并告知运维」的唯一通道。

```bash
# 在仓库根执行：一次性安装到 PATH、写入 /etc/starbyte.conf，
# 并在 deploy/.env 不存在时自动生成随机密钥 + 打印口令
sudo bash deploy/cli/starbyte install
```

之后任意目录：

| 命令 | 作用 |
|------|------|
| `starbyte install` | 装 CLI + 软链 `sb`；`.env` 不存在时自动执行 `init` 并把口令告知运维 |
| `starbyte init [--force]` | 由系统自动生成随机密钥并写 `deploy/.env`。`--force` 会**替换全部密钥项**、保留非密钥键；换库口令后需同步已有数据卷 |
| `starbyte credentials` | 再次显示系统生成的口令（仅限可信终端） |
| `starbyte status` / `sb status` | `compose ps` + `:8080/health`、`:80` 短检查 |
| `starbyte logs [服务] [--tail N]` | 看日志 |
| `starbyte restart [服务\|all]` | 重启容器（不重建镜像） |
| `starbyte rebuild [backend\|frontend\|all]` | `build` + `up --no-deps --force-recreate`，避免顺带强拉 MinIO/Postgres |
| `starbyte env-check` | 检查 `deploy/.env` 关键键名是否齐全、密钥是否仍是占位/弱值；不打印值。**通过返回 0，失败返回 1**，可直接用于 `env-check && compose up` |
| `starbyte backup create\|list\|preview\|restore\|drill` | 经 backend 容器走托管备份（gzip / AES / 预览 / 失败告警） |
| `starbyte backup emergency` | 主机明文 `pg_dump` → `/var/backups/starbyte/`（backend 不可用时的兜底） |
| `starbyte doctor` | 校园部署常见问题：镜像站 403、`SKIP_BUCKET_CREATE`、CAS 回调 / authserver 内网解析、80 端口 |

`rebuild` 对应生产上已验证的绕过方式：镜像站 403 时不要对 MinIO 做 `compose up` 全量拉取。

## 用 zip 更新代码（保留 .env、冻结基础设施、可选强校验）

```bash
# 推荐：附带校验清单
sha256sum StarByte.zip > StarByte.zip.sha256   # 由发布方生成并另行可信分发
bash deploy/update-from-zip.sh /path/to/StarByte-main.zip
starbyte rebuild all

# 偏执模式：只要 zip 里出现任何受保护基础设施文件就中止
bash deploy/update-from-zip.sh --strict /path/to/StarByte-main.zip

# 显式指定哈希（与上面的 <zip>.sha256 二选一，显式优先）
bash deploy/update-from-zip.sh --sha256 <64位hex> /path/to/StarByte-main.zip
```

脚本会把 zip 解到仓库上，**始终保留现有 `deploy/.env`**，且不执行任何 `docker volume` / `down -v`。

**供应链加固（重要行为变更）**：以下基础设施文件/目录**永不**被 zip 覆盖，而是被忽略并在结尾醒目列出「跳过了哪些」——若 zip 里出现它们，说明 zip 可能被伪造或被改动：

- 所有 `Dockerfile` / `Dockerfile.*`（含 `backend/Dockerfile`、`frontend/Dockerfile`）
- `deploy/docker-compose.yml`、`deploy/docker-compose.dev.yml`
- `deploy/cli/**`（整个 CLI 目录，含 `starbyte`）
- `deploy/update-from-zip.sh` 自身及 `deploy/*.sh`
- `backend/entrypoint.sh`、`backend/configs/**`
- `frontend/nginx.conf`、`frontend/security-headers.conf`
- `.github/**`、`Makefile`、`.gitignore`、`.dockerignore`、`backend/.dockerignore`、`frontend/.dockerignore`

因此 `deploy/docker-compose.yml` **不再**会被 zip 覆盖（旧文档说「会被覆盖」已作废）。要更新这些基础设施，请走可信渠道（git 拉取 / 发布方单独下发），不要用微信群 zip 顺带更新——这正是为掐断「微信群 zip → 宿主机 root」的 RCE 链条。

**完整性校验**：脚本会自动探测同目录的 `<zip>.sha256`（兼容 `sha256sum` 输出 `<hex>  <文件名>` 与纯 hex 行）；提供则强校验，不匹配即中止、**不写入任何文件**。未提供则打印醒目警告后继续（救火场景不硬失败，但风险自担）。发布方生成清单：

```bash
sha256sum StarByte.zip > StarByte.zip.sha256
```

CAS 所需的 `extra_hosts` 已写进仓库 compose（受保护、不会被 zip 改动），`rebuild` 即可保持 `authserver.smbu.edu.cn` → `CAS_AUTHSERVER_IP`。

## 手动部署

1. PostgreSQL 16、Redis 7、MinIO。
2. 后端：`cd backend && go build -o server ./cmd/server`（**需 Go ≥ 1.25**，见 `go.mod`），设置 `APP_ENV=prod`、`CONFIG_PATH` 与 `DB_*` / `REDIS_*` / `MINIO_*` / `JWT_SECRET`。
   口令同样应走随机生成，不要手填：可用 `starbyte init` 生成 `deploy/.env` 后从中取值，
   或自行用 `openssl rand -base64 48` 一类方式生成 32 位以上高熵串。
   生产 `JWT_SECRET` 若长度 < 32、命中占位词表、或字符种类 ≤ 2，**后端会拒绝启动**。
3. 迁移：`make migrate-up DATABASE_URL=postgres://...`。
4. 种子：`APP_ENV=prod make seed`（admin 口令生成与抄录方式见上）。
5. 前端：`cd frontend && npm ci && npm run build`，用 Nginx 托管 `dist/` 并把 `/api/` 反代到后端。
   反代配置请参考 `frontend/nginx.conf`（含安全响应头与 HTTPS 模板），不要只写裸 `proxy_pass`。

## 学校统一认证（漏测先用 IP）

- 登录页「学校统一认证」跳到 `https://authserver.smbu.edu.cn/authserver/login`
- **漏测没有域名**：用校园网 IP 打开系统（`http://<服务器IP>/`）。`service` / 回跳按访问 Host 自动拼，信息化备案：
  `http://<服务器IP>/api/v1/auth/cas/callback`
- 有 `starbyte.smbu.edu.cn` 后再改备案，或设 `CAS_SERVICE_URL` / `CAS_FRONTEND_URL`
- **容器解析**：backend 带 `extra_hosts`，`authserver.smbu.edu.cn` → `${CAS_AUTHSERVER_IP:-<compose 内置默认值>}`。校园 DNS 常返回不可达 IPv6，callback 会 wget 超时约 20s（nginx 502）。改 IP 后必须 `starbyte rebuild backend`（或 `--force-recreate`）才会写入容器 `/etc/hosts`
- 具体内网地址只写在运维本机 `deploy/.env`，不提交回仓库（`deploy/.env.example` 用 `<...>` 占位符，`starbyte init` 会忽略占位符并让 compose 默认值兜底）
- 环境变量见 `deploy/.env.example` / `backend/.env.example` 的 `CAS_*`
- 外网 `starbyte.com` 检测校内 IP 后 301 属二期

## Nginx 反向代理（示例）

**优先直接用仓库里已加固的 `frontend/nginx.conf`**（由 `frontend/Dockerfile` 打进镜像），
它已包含：`server_tokens off`、`nosniff` / `X-Frame-Options` / `Referrer-Policy` /
`Permissions-Policy` / CSP、非 root 监听 8080、以及 443 server 块模板。

若你在宿主机另起 nginx（不用容器内的那个），至少要照抄安全头：

```nginx
server {
    listen 443 ssl;
    http2 on;
    server_name _;
    ssl_certificate     /etc/nginx/certs/starbyte.crt;
    ssl_certificate_key /etc/nginx/certs/starbyte.key;
    ssl_protocols       TLSv1.2 TLSv1.3;

    server_tokens off;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    root /var/www/starbyte/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

# 80 → 443
server {
    listen 80;
    server_name _;
    return 301 https://$host$request_uri;
}
```

## HTTPS（上线前必须完成）

**纯 HTTP + IP 只允许作为临时调试手段，不得作为正式上线形态。**
理由：登录口令与 Bearer token 在校园局域网内明文传输，同网段可用 ARP 欺骗 / 流量镜像窃取；
一旦配合固定或弱 JWT 密钥，攻击者可直接离线签发超管 token。

落地步骤：

1. 取证书：校园 CA 签发，或自签（自签需把 CA 证书分发给客户端，否则浏览器告警）。
2. 把证书放到宿主机 `deploy/certs/starbyte.crt` 与 `starbyte.key`。
3. 打开 `frontend/nginx.conf` 末尾的 443 server 块与 80→443 跳转块（取消注释）。
4. 在 `deploy/docker-compose.yml` 里打开 `443:443` 端口映射与 `./certs:/etc/nginx/certs:ro` 卷挂载。
5. `starbyte rebuild frontend`。
6. 验证：`curl -sI https://<地址>/ | grep -i strict-transport` 应看到 HSTS；浏览器地址栏应为锁标。
   HSTS 首次建议先用小 `max-age`（如 300）验证无混合内容问题，再调大。

不要单独做 `auth.` 子域，CAS 回调与站点同 Host：`https://<IP或域名>/api/v1/auth/cas/callback`。

## 上线前核查清单

- [ ] `deploy/.env` 由 `starbyte init` 生成，**没有任何一项是占位/弱值**（`starbyte env-check` 通过）
- [ ] 安装时打印的 6 项密钥已抄录进密码管理器（终端滚动缓冲会被清掉，别只依赖它）
- [ ] `starbyte env-check` 返回 0（失败会返回非 0，可安全用于 `&&` 链）
- [ ] `deploy/.env` 与 `deploy/.env.credentials` **未被提交**（`git check-ignore -v deploy/.env`）
- [ ] `APP_ENV=prod`（`deploy/docker-compose.yml` 已写死；手动部署时自查）
- [ ] `METRICS_TOKEN` 已设或明确接受 `/metrics` 返回 404
- [ ] HTTPS 已启用，且 `Strict-Transport-Security` 生效
- [ ] `make seed` 的 admin 随机口令已抄录、已登录改密
- [ ] MinIO 桶 `starbyte` 已手动创建
- [ ] 后端 8080 / MinIO 9000 **未**对校园网开放（`ss -lnt` 确认只绑 127.0.0.1）
- [ ] 已确认无 `down -v` / 无删除 named volumes 的操作习惯

## 数据库备份与恢复

```bash
# 推荐：与应用同一套 custom dump + gzip + 可选 AES
starbyte backup create
starbyte backup preview <id>
# 演练到独立库（先 createdb starbyte_drill；不是 PITR）
# 同集群换库名可省略密码；换主机须 --password 或 DSN 内密码
starbyte backup drill <id> --dbname starbyte_drill --confirm DRILL
starbyte backup drill <id> --dbname stolen --host other.example --password "$DRILL_PASSWORD" --confirm DRILL

# backend 挂了才用明文应急
starbyte backup emergency

# 覆盖当前应用库（危险）
starbyte backup restore <id> --confirm RESTORE
```

WAL / PITR 不在应用内：需要数据库主机配置 `wal_level`、归档命令和 `pg_basebackup`。

MinIO 桶与 Postgres 一起备份。Named volumes 典型路径可用 `docker volume inspect` 查看（compose 项目名多为 `deploy`，卷名类似 `deploy_postgres-data`）。恢复后确认迁移版本；`SKIP_BUCKET_CREATE=true` 时桶需已存在。
