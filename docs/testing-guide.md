# StarByte 测试规范

> 对应 Issue #30。后端 `xxx_test.go`，前端 `xxx.test.ts(x)`，结构一律 Arrange-Act-Assert。

## 1. 后端

- 框架：`testing` + `testify` + `httptest` + `go.uber.org/mock`（gomock）
- 集成库：`pkg/testutil`（Gin JSON 请求、`OpenPostgres` 连接测试库）
  - 默认：连不上 PostgreSQL 则 `t.Skip`
  - CI：`TEST_DATABASE_REQUIRED=1`，连不上则失败
- Mock 生成：在接口文件加 `//go:generate mockgen ...`，执行 `go generate ./internal/rbac/repo`
- 覆盖率目标（长期）：核心模块（auth / user / rbac / workflow）>80%，业务模块 >60%，service 层 >80%
- CI 硬门禁：对**当前已达标**的包失败即阻止合并（见 `backend/scripts/check-coverage.sh`）
  - 始终：`internal/auth/repo`、`internal/rbac`、`internal/rbac/model` ≥80%
  - CI 有测试库时额外：`internal/auth/service` ≥80%（gomock 生成文件不计入覆盖率）

```bash
cd backend
go test ./...
make backend-test
make backend-cover          # 覆盖率 + 核心包门禁
TEST_DATABASE_URL='postgres://starbyte:starbyte@localhost:5432/starbyte_test?sslmode=disable' \
  TEST_DATABASE_REQUIRED=1 go test ./internal/workflow/repo ./internal/rbac/repo ./internal/stats/repo
```

Handler 集成测试示例：

```go
r := testutil.NewEngine()
r.POST("/users", h.CreateUser)
w := testutil.JSONRequest(t, r, http.MethodPost, "/users", req, nil)
require.Equal(t, 200, w.Code)
```

Service 层用 gomock 隔离 repo：

```go
ctrl := gomock.NewController(t)
mockRepo := repo.NewMockPermissionRepo(ctrl)
svc := NewPermissionService(nil, mockRepo, nil)
mockRepo.EXPECT().GetByCode(gomock.Any(), "user:read").Return(nil, nil)
```

## 2. 前端

- 框架：Vitest + Testing Library（jsdom）
- E2E：Playwright 预留，本仓库暂不强制安装/运行
- 覆盖率：`npm run test:coverage`（当前纳入 FormEngine / StatusTag / EmptyState / DataTable）

```bash
cd frontend
npm test -- --run
npm run test:coverage
```

组件测试覆盖核心交互（渲染、搜索、条件显示），文件名 `ComponentName.test.tsx`。

## 3. CI

`.github/workflows/ci.yml`：

- 后端：PostgreSQL 16 service + 迁移后 `go test -race ./...`，失败阻止合并
- 覆盖率：`scripts/check-coverage.sh`，产物 `backend/coverage.out` 作为 artifact
- 前端：`npm run test:coverage`，失败阻止合并；产物 `frontend/coverage/`

本地无 Docker 时，依赖数据库的测试会 Skip，不影响单测；覆盖率门禁里需要库的包仅在 CI 强制。
