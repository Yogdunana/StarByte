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

- 入会：`/member/applications`、审核 approve/reject/supplement、`/member/profiles`
- 面试：`/interviews`、sessions、evaluations、stats
- 会议：`/meetings`、attendees、agendas、votes
- 任务：`/tasks`、指派/转交/评论/附件、`/tasks/my/*`
- 实习：`/internships`、complete/report、stats

### 流程 / 表单 / 文件 / 通知 / 统计

- 流程定义与实例：`/workflow/definitions`、`/workflow/instances`、`/workflow/tasks`
- 表单：`/forms`、submit、submissions
- 文件：`POST /files/upload`、`GET /files/:id/download`
- 通知：`/notifications`、模板、邮件
- 统计：`/stats/overview`、`/stats/{provider}`、export

### 系统

- RBAC：`/system/roles` `/system/permissions` `/system/departments` `/system/positions`
- 审计：`/system/audit-logs`（含 traces/reports/archives）
- 字典 / 配置 / 缓存 / 调度 / 搜索 / 导出：见 `/system/*` 与 `/export/*`

健康检查（无前缀）：`GET /health` `GET /health/ready` `GET /metrics`。
