# API 文档

完整请求/响应以运行中的 OpenAPI 为准。

- 非生产后端启动后：http://localhost:8080/swagger/index.html
- 规范由 Handler 注释生成：`make swagger`（`backend/scripts/swagger.sh`）
- 统一信封：`{ code, message, data, request_id, timestamp }`
- `code = 0` 表示成功；分页在 `data.list / total / page / page_size`

## 错误码分段

见 `TEAM_DEV_GUIDE.md` §3.4 与 `backend/pkg/response/error_codes.go`。

常用段：

| 范围 | 模块 |
|------|------|
| 1000-1999 | 通用（1001 参数、1002 未登录、1501 预留未实现） |
| 2000-2999 | 用户/认证 |
| 3000-3999 | RBAC |
| 4000-4999 | 流程引擎 |
| 23000-23999 | 财务 |
| 24000-24999 | 纪律处分 |
| 25000-25999 | 合同 |

财务 `GET /api/v1/finance/export` 一期返回 **501 / 1501**，二期再接导出引擎。
