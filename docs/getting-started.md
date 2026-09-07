# 开发者入门

目标：30 分钟内在本机跑起前后端。

## 环境准备

- Docker + Docker Compose
- Go 1.22+
- Node.js 18+
- 可选：`golang-migrate` CLI（`make migrate-up`）

## 首次运行

```bash
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte

# 1. 基础设施（Postgres / Redis / MinIO，端口打到主机）
docker compose -f deploy/docker-compose.dev.yml up -d

# 2. 迁移（库名与 config.dev.yaml 对齐：starbyte_dev）
make migrate-up POSTGRES_DB=starbyte_dev

# 3. 种子（Makefile 默认 APP_ENV=dev，写入 starbyte_dev）
make seed

# 4. 后端（必须编译整个 cmd/server 包）
cd backend
APP_ENV=dev go run ./cmd/server

# 5. 另开终端：前端
cd frontend
npm install
npm run dev
```

打开 http://localhost:5173 ，用 `admin/admin123` 登录（也可用学号 `20210001`）。

不要只跑 `go run cmd/server/main.go`：同包还有 swagger / traffic 文件，会缺符号。

## 常见问题

| 现象 | 处理 |
|------|------|
| 连不上 Postgres | 确认用的是 **dev** compose；生产 compose **不暴露** 5432 |
| 迁移失败 | `POSTGRES_DB` 与 `backend/configs/config.dev.yaml` 一致 |
| 登录 429 | #75 用户令牌桶约 2 req/s；稍等再试 |
| 菜单没有财务/处分/合同 | 重新 `make seed` 写入权限；超管/社长默认全开 |
| Swagger 404 | 仅非生产启用；访问 `/swagger/index.html` |
| `any` / 单文件超 300 行 | 见 `TEAM_DEV_GUIDE.md`，CI 会拦 |

## 代码规范摘要

- 分层：handler → service → repo → model
- 禁止 TypeScript `any`
- 错误码按模块分段（`TEAM_DEV_GUIDE.md` §3.4）
- 单测：`make backend-test`、`cd frontend && npm test -- --run`

## 提交第一个 PR

1. 从最新 `main` 拉分支：`git checkout -b feature/xxx`
2. 认领对应 Issue（评论格式见 Issue 模板）
3. 开发、补测、自测主路径
4. 开 PR，至少 1 人 Review，Squash Merge
5. 详细流程：`docs/dev-guide/git-workflow.md`、`docs/dev-guide/pr-specification.md`
