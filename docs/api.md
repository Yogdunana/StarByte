# API 文档

前缀：`/api/v1`。完整请求/响应以运行中的 OpenAPI 为准。

- 非生产：http://localhost:8080/swagger/index.html
- 从 Handler 注释生成：`make swagger`
- 统一信封：`{ code, message, data, request_id, timestamp }`
- `code = 0` 表示成功；分页在 `data.list / total / page / page_size`
- 鉴权：`Authorization: Bearer <access_token>`（登录接口除外）

## 错误码分段

见 `TEAM_DEV_GUIDE.md` §3.4 与 `backend/pkg/response/error_codes.go`。

| 范围 | 模块 |
|------|------|
| 1000-1999 | 通用（1001 参数、1002 未登录、1501 预留未实现） |
| 2000-2999 | 用户/认证 |
| 3000-3999 | RBAC |
| 4000-4999 | 流程引擎 |
| 23000-23999 | 财务 |
| 24000-24999 | 纪律处分 |
| 25000-25999 | 合同 |
| 26000-26999 | 值班（预留） |
| 27000-27999 | 活动 |
| 28000-28999 | 公告 |
| 29000-29999 | 日程 |
| 30000-30999 | 请假 |
| 31000-31999 | 运维监控 |
| 32000-32999 | 数据备份 |
| 33000-33999 | 知识库 / CMS |
| 34000-34999 | 特性开关 / 灰度 |

## 请求 / 响应示例

登录：

```http
POST /api/v1/auth/login
Content-Type: application/json

{"username":"admin","password":"admin123"}
```

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<jwt>",
    "expires_in": 7200
  }
}
```

新增财务记录：

```http
POST /api/v1/finance/records
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "秋季会费",
  "category_id": "<uuid>",
  "direction": 2,
  "amount": 50,
  "occurred_at": "2026-09-07T00:00:00Z"
}
```

`direction`：`1` 支出、`2` 收入，**必须与分类的方向一致**（否则 `23005`）。`GET /finance/export` 一期返回 **501 / 1501**，二期再接导出引擎。

登记处分：

```http
POST /api/v1/discipline/records
{"user_id":"<uuid>","title":"警告","level":1,"description":"迟到"}
```

`level`：1 警告 … 5 开除会籍。`status`：0 待审批、1 生效、2 已撤销、3 申诉中。

新增合同：

```http
POST /api/v1/contracts
{
  "title": "赞助协议",
  "party_name": "某公司",
  "contract_type": 1,
  "amount": 5000,
  "start_at": "2026-09-01",
  "expired_at": "2026-12-31",
  "file_id": "<uuid from POST /files/upload>"
}
```

`contract_type`：1 赞助、2 活动、3 采购、4 其他。`POST` 一律创建为草稿（`status=0`），忽略请求体中的 `status`；生效/终止走 `PUT` 且需要 `contract:manage`。`status`：0 草稿、1 生效中、2 已到期、3 已终止。

## 接口列表

路径均相对 `/api/v1`。

### 认证 / 用户

- `POST /auth/login` `POST /auth/register` `POST /auth/refresh` `POST /auth/logout`
- `GET /auth/me` `PUT /auth/password`
- `GET /auth/sessions` 及强制下线；OAuth/企微接口预留
- 学校 CAS（校园网 ↔ `authserver.smbu.edu.cn`）
  - `GET /auth/cas/status` 是否开通
  - `GET /auth/cas/login?redirect=/dashboard` 302 到金智 `/authserver/login`
  - `GET /auth/cas/callback?ticket=` 验 ST 后 302 到 `/login/cas?code=`（state 走 Cookie，callback 路径固定）
  - `POST /auth/cas/exchange` `{ "code" }` 换本系统 JWT（复用 #17 签发）
  - 漏测无域名：用浏览器访问的 IP 拼 `http://<IP>/api/v1/auth/cas/callback` 给信息化备案
  - 有域名后可设 `CAS_SERVICE_URL` / `CAS_FRONTEND_URL`，或继续留空跟访问地址走
- `GET /users` `POST /users` `GET|PUT|DELETE /users/:id`
- `GET /user/me` `PUT /user/profile` `PUT /user/password`

### 财务 #22

- `GET|POST /finance/records` `GET|PUT|DELETE /finance/records/:id`
- `GET /finance/categories` `GET /finance/summary` `GET /finance/export`（501）

### 纪律处分 #23

- `GET|POST /discipline/records` `GET|PUT /discipline/records/:id`
- `POST /discipline/records/:id/approve|revoke|appeal`

### 合同 #24

- `GET|POST /contracts` `GET|PUT|DELETE /contracts/:id`
- `GET /contracts/templates` `GET /contracts/expiring`

### 会员 / 面试 / 会议 / 任务 / 实习

- 入会：`/member/applications`、审核 approve/reject/supplement/transfer（新申请走可配置 `member_application`：干事→部长→社长，`GET .../progress` 看进度）、`/member/profiles`
- 面试：`/interviews`、sessions、evaluations、stats
- 会议：`/meetings`、attendees、agendas、votes
- 任务：`/tasks`、指派/转交/评论/附件、`/tasks/my/*`；开启审核验收的新任务走 `task_lifecycle` 流程实例（`GET|POST /tasks/:id/workflow`，拒绝终止实例）。委托/转办：`POST /tasks/:id/transfer` 或 `POST /tasks/:id/handover`（同部门直接委托，跨部门/中心走 `task_transfer_*` 签字）；`GET /tasks/:id/handover`、`POST /tasks/:id/handover/decisions`（body 须带 `transfer_id`）、`GET|POST /tasks/transfers/:id`。创建时可按部门/角色/轮询自动分配。超时由调度扫描流程待办 `dueDays` 并升级/重新分配。
- 实习：`/internships`、complete/report、stats
- 请假：`GET|POST /leave`、`GET /leave/my`、`GET /leave/types`、`POST /leave/types`、`PUT /leave/types/:id`、`GET /leave/balance`、`GET /leave/stats`、`GET /leave/calendar`、`GET /leave/todos`、`PUT /leave/:id/approve|reject`；新申请走 `leave_approval` 流程实例（部长→社长），历史记录无实例时单级回退

### 流程 / 表单 / 文件 / 通知 / 统计

- 流程定义与实例：`/workflow/definitions`、`/workflow/instances`、`/workflow/tasks`；任务转办待办 `business_type=task_transfer` 须在任务面板签字，不能走通用 CompleteTask
- 表单：`/forms`、submit、submissions
- 文件：`POST /files/upload`、`GET /files/:id/download`
- 通知：`/notifications`、模板、邮件
- 统计：`/stats/overview`、`/stats/{provider}`、export

### 系统

- RBAC：`/system/roles` `/system/permissions` `/system/departments` `/system/positions`
- 审计：`/system/audit-logs`（含 traces/reports/archives）
- 字典 / 配置 / 缓存 / 调度 / 搜索 / 导出：见 `/system/*` 与 `/export/*`

健康检查（无前缀）：`GET /health` `GET /health/ready` `GET /metrics`。

### 运维监控（#87，`monitor:read`）

- `GET /monitor/server` CPU / 内存 / 磁盘 / 负载
- `GET /monitor/app` 运行时长、Goroutine、MemStats / GC
- `GET /monitor/database` `sql.DB` 连接池
- `GET /monitor/redis` INFO（连接数 / 内存 / 命中率，不含主机凭据）
- `GET /monitor/api-stats` Prometheus 请求计数 + P50/P95/P99（最近请求窗口，不足时回退直方图插值）
- `GET /monitor/slow-queries` 慢查询：`pg_stat_statements`（若已安装）否则 `pg_stat_activity` + 进程内 GORM 环；附 Redis SLOWLOG
- `WS /ws/monitor` 实时快照推送（JWT query `token` 或 Bearer；需 `monitor:read`）。前端轮询为回退。

进程列表 / 网卡流量深挖本切片不做。

### 数据备份（#88，运维角色）

- `GET /system/backups` 备份列表（`backup:read`）
- `POST /system/backups` 手动触发全量 `pg_dump --format=custom` + gzip（`.dump.gz`），异步落 MinIO（`backup:create`）
- `GET /system/backups/:id` 备份详情（校验和 / 大小 / 状态）
- `DELETE /system/backups/:id` 删除对象与记录（`backup:delete`）
- `POST /system/backups/:id/restore` 恢复到**当前应用库**；请求体须 `confirm=true` 且 `confirmation=RESTORE`（`backup:restore`）。成功/已恢复/恢复失败（2/5/6）可发起。底层为 `pg_restore --single-transaction --clean --if-exists`。
- `POST /system/backups/:id/restore-drill` 恢复演练到**独立 Postgres**（`confirmation=DRILL` + `target_dbname` 和/或 `target_dsn`）。立即返回 `{ queued: true }`，不改生产库记录状态。拒绝指向当前应用库（含 compose 主机别名 `postgres` / `starbyte-postgres` / `starbyte-postgres-dev`）。`target_dbname` 只能是普通标识符（拒绝 `=` / `postgres://`，避免被 `pg_restore -d` 当成 conninfo）。换主机须提供 `target_password` 或 DSN 密码，**不会**复用生产库密码。**不是 PITR**。
- `GET /system/backups/:id/restore-drill` 轮询演练结果（`backup:restore`）。内存态，进程重启后需重做。
- `GET /system/backups/:id/preview` 完整性检查（SHA-256 / 解密 / gzip / TOC），不执行恢复。
- `GET|PUT /system/backups/policies` 保留天数 + 6 字段 cron（`backup:read` / `backup:manage`）；调度同步失败时接口报错，不假装成功。
- `GET /system/backups/storage` 成功 / 已恢复 / 恢复失败备份条数与体积；`encryption_enabled` / `incremental_enabled=false` / `pitr_enabled=false`

CLI：`starbyte backup create|preview|restore|drill` 调用容器内 `starbyte-server backup …`，与 HTTP 同一套 gzip / AES-256 / 告警。`starbyte backup emergency` 才是主机明文 `pg_dump`。

WAL 增量 / 指定时间点恢复（PITR）需要主机级 `pg_basebackup` + WAL 归档，本模块是逻辑 `pg_dump -Fc`，不假装支持。

### 知识库 / CMS（#58）

公开（可选 JWT；`visibility=public` 且已发布无需登录）：

- `GET /knowledge/public/pages/:slug` 独立页（如 `about-us`）
- `GET /knowledge/public/docs` 文档列表
- `GET /knowledge/public/docs/:slug` 文档正文；成员手册未登录返回 33004
- `GET /knowledge/public/tree` 可见目录树
- `GET /knowledge/public/search?q=` 全文搜索（PostgreSQL FTS + ILIKE）

管理（需 `doc:read` / `doc:create` / `doc:update` / `doc:delete` / `doc:publish`）：

- `GET|POST /knowledge/docs`、`GET|PUT|DELETE /knowledge/docs/:id`
- `POST /knowledge/docs/:id/publish`
- `GET /knowledge/docs/:id/history`、`GET /knowledge/docs/:id/versions/:version`
- `POST /knowledge/docs/:id/rollback` 回滚并生成新版本
- `GET /knowledge/search`、`GET /knowledge/tree`、分类 CRUD
- `POST /knowledge/docs/:id/attachments` 关联已有 `files` 记录

公告模块（#77）仍是时效通知，不并入知识库。

### 特性开关 / 灰度（#98）

管理（`feature:read` / `feature:create` / `feature:update` / `feature:manage`）：

- `GET|POST /system/features`、`GET|PUT /system/features/:id`
- `GET /system/features/:id/evaluate` 按用户评估（可带 `user_id`）
- `GET /system/features/:id/analytics?days=7` 独立用户曝光（按变体）
- `POST /system/features/:id/toggle` 切换启用
- `POST /system/features/:id/rollback` 回滚最近一次可逆审计（无快照 `34008`）
- `GET /system/features/audit`

公开 SDK：

- `GET /features/me?keys=cms.public,announcement.feed`（可选 JWT；名单/百分比/AB 匿名 fail closed）

规则字段：`starts_at` / `ends_at`（定时）、`environments`（`dev|test|prod`）、`variants`（`ab_test`）。WASM 客户端不做，浏览器走 HTTP SDK。
