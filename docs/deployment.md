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
| Swagger | 默认 **404**（fail-closed：只有显式 `APP_ENV=dev\|test` 才挂载）。本地开发见 [getting-started.md](getting-started.md)；未带 `APP_ENV` 裸跑二进制会 404，属预期行为 |
| MinIO API | 仅本机 `http://127.0.0.1:9000`；控制台请用 `ssh -L 9001:127.0.0.1:9001` 转发后访问 |

首次启动后端会跑迁移，**但不会跑种子**（`entrypoint.sh` 只做迁移、建桶、起服务）。
生产 Postgres **不映射主机端口**，在宿主机直接 `APP_ENV=prod make seed` 会连不上库。
compose 部署请用 `starbyte bootstrap`（它会起一个加入 compose 网络的一次性容器，
并自动带上 `DB_HOST` / `DB_USER` / `DB_PASSWORD` / `JWT_SECRET` 等，见下节）。

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
- 开发便利口令（`admin/admin123`、`test/test123`）**只在显式 `APP_ENV=dev|test` 时写入**
  （见上）。`APP_ENV` 未设置同样是生产口径，不会写入开发口令 —— 不要按「非 prod 就是非生产」
  来理解，空值按生产处理。

### compose 部署下怎么跑种子

容器化部署**不会自动跑种子**：`backend/entrypoint.sh` 只做「迁移 → MinIO 建桶 → 起服务」，
生产镜像里也没有 Go 工具链。管理员要显式建：

```bash
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --build
starbyte bootstrap          # 写入种子 + 创建 admin，初始口令打印一次
```

`starbyte bootstrap` 走一次性容器（复用 `backend` 服务定义、只覆盖 entrypoint，
不新增 compose 服务，因此默认 `up -d` 的行为不变）。它是幂等的：admin 已存在就跳过，
不回显既有口令。

初始口令**只显示一次**，丢了也不需要重来：

```bash
starbyte admin show                     # 查看管理员账号（默认列出全部 super_admin）
starbyte admin set-password             # 重设口令，不需要旧口令
starbyte admin set-password --value '…' # 按指定值设置（会进 shell history，一般不推荐）
starbyte admin set-username <新账号名>   # 改账号名
```

三条 `admin` 命令都会吊销该账号的全部在线会话与 refresh token，并清掉失败计数与登录锁定。
改过账号名后，后续命令要带 `-u <新账号名>`。

> 为什么不能只改数据库：核心角色（`super_admin` / `president` 等）**不在迁移里**，
> 全库只有 `000072` / `000073` / `000077` 三条迁移写 `INSERT INTO roles`，且都是后加的
> 章程角色；角色本体在 `backend/scripts/seed_rbac.go`。手工 `INSERT` 一个用户的话，
> `user_roles` 的 `CROSS JOIN roles r WHERE r.code='super_admin'` 匹配不到任何行、
> **不报错但插 0 行**，结果是一个能登录却没有任何权限的账号。

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
| `starbyte bootstrap` | 写入种子并创建管理员账号（幂等）。栈没起会自动拉起 postgres/redis/minio |
| `starbyte admin show` | 查看管理员账号 |
| `starbyte admin set-password [--value …]` | 重设管理员口令（不需要旧口令），并吊销其全部在线会话 |
| `starbyte admin set-username <新账号名>` | 改管理员账号名 |
| `starbyte status` / `sb status` | `compose ps` + `:8080/health`、`:80` 短检查 |
| `starbyte logs [服务] [--tail N]` | 看日志 |
| `starbyte restart [服务\|all]` | 重启容器（不重建镜像） |
| `starbyte rebuild [backend\|frontend\|all]` | `build` + `up --no-deps --force-recreate`，避免顺带强拉 MinIO/Postgres |
| `starbyte env-check` | 检查 `deploy/.env` 关键键名是否齐全、密钥是否仍是占位/弱值；不打印值。**通过返回 0，失败返回 1**，可直接用于 `env-check && compose up`。拓扑项不强制填，但**填了就校验合法性**（`TRUSTED_PROXIES` 必须是 CIDR） |
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
   反代配置请参考 `frontend/nginx.conf`（含安全响应头），不要只写裸 `proxy_pass`。

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
`Permissions-Policy` / CSP、非 root 监听 8080。
HTTPS 的 443 server 块**不在这里**，而在 `deploy/nginx/starbyte-https.conf.template`，
由 `starbyte ssl enable` 挂载生效（见「HTTPS」一节）。

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
    # HSTS 首次用 300 秒验证无混合内容后再调大；大 max-age 一旦下发很难回退。
    add_header Strict-Transport-Security "max-age=300" always;
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

# 80 端口：**默认不做全量 301**。
# 强制跳转叠加 HSTS 大 max-age，会把一次证书配置错误变成不可恢复的锁定；
# 校园网内用 IP 直访、CAS 回调、健康检查也都可能依赖明文 HTTP。
# 容器内那份配置（frontend/nginx.conf）就是不做跳转的。
# 若你确实要跳，至少给 ACME 挑战留一条明文通道：
server {
    listen 80;
    server_name _;

    location ^~ /.well-known/acme-challenge/ {
        root /var/www/acme;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}
```

## HTTPS（上线前必须完成）

**纯 HTTP + IP 只允许作为临时调试手段，不得作为正式上线形态。**
理由：登录口令与 Bearer token 在校园局域网内明文传输，同网段可用 ARP 欺骗 / 流量镜像窃取；
一旦配合固定或弱 JWT 密钥，攻击者可直接离线签发超管 token。

### 用 CLI 开启（默认路径）

一期由学校提供 `*.smbu.edu.cn` 通配证书，整件事就是三条命令：

```bash
# 1. 体检 + 启用（校验证书与私钥是否配对、是否过期、是否覆盖目标域名）
sudo starbyte ssl enable --cert /root/starbyte.smbu.edu.cn.crt \
                         --key  /root/starbyte.smbu.edu.cn.key \
                         --host starbyte.smbu.edu.cn

# 2. 看状态（签发者 / 到期日 / 剩余天数 / 是否已启用）
starbyte ssl status

# 3. 出问题时一键退回纯 HTTP（逃生门）
starbyte ssl disable
```

只想体检不想动运行状态：`starbyte ssl verify --cert <crt> --key <key> --host <域名>`。

`enable` 依次做：把证书复制到 `deploy/certs/starbyte.{crt,key}`（私钥 0600）→
从 `deploy/nginx/starbyte-https.conf.template` 生成
`deploy/nginx/https.d/starbyte-https.conf` → 构建 frontend →
**用一次性容器跑 `nginx -t` 预检** → 通过才重建容器。
预检不通过会连同证书文件一起自动回滚，**不会把站点搞挂**。

### 机制（改动前先读）

- `frontend/nginx.conf` 末尾只有一行通配 include：
  `include /etc/nginx/snippets/https.d/*.conf;`。
  通配 include 在目录为空时匹配不到文件且**不报错** —— 未启用时就是纯 HTTP，
  行为与引入本机制之前完全一致。
- 启用状态的唯一事实来源 = `deploy/nginx/https.d/starbyte-https.conf` 是否存在。
  CLI 的 `compose()` 据此自动叠加 `deploy/docker-compose.https.yml`。
  **只在启用时才叠加**：否则宿主 443 被占用会让最普通的 `docker compose up` 直接失败。
- **不要手工去注释 compose / nginx.conf。** 那样既不幂等（跑第二次就坏），
  也绕过了证书配对校验与自动回滚。

### 两件必须注意的事

1. **不做 80→443 强制跳转。** HTTP 入口始终保留 —— 那是证书配错或过期时的逃生通道。
   强制跳转叠加 HSTS 大 `max-age`，会把一次配置错误变成不可恢复的锁定。
   强制走 HTTPS 交给 HSTS 在浏览器侧完成。
2. **HSTS 首次用 `max-age=300`。** 确认无混合内容问题后再逐步调大；一旦下发大值，
   浏览器会在有效期内强制 HTTPS，想退回 HTTP 很痛苦。

### 证书来源怎么选

- **一期（学校通配证书，推荐）**：向信息化要 **PEM 格式证书 + 私钥 + 中间证书链**，缺一不可。
  不少单位习惯只发 crt 不给 key（私钥不该离席是正当顾虑），那样部署不了，开口时就说清楚。
- 注意 `*.smbu.edu.cn` 的星号**只覆盖一级子域**，不覆盖裸域 `smbu.edu.cn`。
  所以入口必须是 `starbyte.smbu.edu.cn` 这样的子域名，不能用裸域。
- **自签**：只用于学校证书到位前先把整条链路联调通 ——
  `starbyte ssl selfsign --host <域名> --yes`。
  不要用于上线：每个同学首次访问都会看到红色告警页，微信里点开还会被拦。
- **Let's Encrypt 一期用不上。** 将来若需要（例如二期 `starbyte.work`）：
  单域名证书走 HTTP-01 即可；通配证书必须走 DNS-01，且每 90 天要改一次
  `_acme-challenge` 的 TXT，必须自动化，人肉续期会累死。

不要单独做 `auth.` 子域，CAS 回调与站点同 Host：`https://<域名>/api/v1/auth/cas/callback`。

## 上线前核查清单

- [ ] `deploy/.env` 由 `starbyte init` 生成，**没有任何一项是占位/弱值**（`starbyte env-check` 通过）
- [ ] 安装时打印的 6 项密钥已抄录进密码管理器（终端滚动缓冲会被清掉，别只依赖它）
- [ ] `starbyte env-check` 返回 0（失败会返回非 0，可安全用于 `&&` 链）
- [ ] `deploy/.env` 与 `deploy/.env.credentials` **未被提交**（`git check-ignore -v deploy/.env`）
- [ ] `APP_ENV=prod`（`deploy/docker-compose.yml` 已写死；手动部署时自查）
- [ ] `METRICS_TOKEN` 已设或明确接受 `/metrics` 返回 404
- [ ] HTTPS 已启用且证书有效（`starbyte ssl status` 返回 0），`Strict-Transport-Security` 生效
- [ ] 证书到期日已记录（`starbyte ssl status` 会在剩余 < 30 天时告警）
- [ ] `starbyte bootstrap` 的 admin 初始口令已抄录、已登录改密（`starbyte admin show` 可复查账号）
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
