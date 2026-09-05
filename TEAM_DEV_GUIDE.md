# StarByte 团队开发规范总纲

> **本文档是 StarByte 项目唯一的一站式开发参考。**
> **所有团队成员（含 AI 助手）在开始任何开发工作前，必须先阅读本文档。**
>
> **新手快速上手 → [DEV_QUICKSTART.md](DEV_QUICKSTART.md)**（一句话认领任务 + AI 提示词模板）
>
> 仓库地址：https://github.com/Yogdunana/StarByte
> 最后更新：2026-09-04

---

## 一、项目概述

StarByte 是面向高校计算机协会的一体化管理平台。Monorepo 架构，Go 后端 + React 前端，支持 20+ 人并行协作开发。

### 核心功能

| 模块 | 说明 |
|------|------|
| 入会申请 | 学生申请会员/干事，干事需经一面/二面，面试流程可拖拽配置 |
| 人员档案 | 用户全量信息管理、变更历史 |
| 面试管理 | 面试安排、评分、多轮面试 |
| 会议投票 | 等权投票（匿名）+ 加权投票（按职务赋权） |
| 任务流转 | 任务创建、分配、跟踪、状态流转 |
| IT 实习管理 | 实习时间段、时长统计、排名 |
| 流程引擎 | 可拖拽可视化流程设计器，插件化节点架构 |
| 数据统计 | ECharts 可视化报表 |
| RBAC 权限 | 角色/权限/部门/职位，数据权限范围 |
| 消息通知 | 站内消息 + WebSocket + 邮件 |
| 审计日志 | 全写操作记录 |

### 技术栈

| 层面 | 技术 | 版本 |
|------|------|------|
| 后端语言 | Go | 1.22 |
| 后端框架 | Gin | - |
| ORM | GORM | - |
| 数据库 | PostgreSQL | 16 |
| 缓存 | Redis | 7 |
| 对象存储 | MinIO | - |
| 前端语言 | TypeScript | 5 |
| 前端框架 | React | 18 |
| 构建工具 | Vite | 5 |
| 状态管理 | Redux Toolkit | 2 |
| UI 组件库 | Ant Design | 5 |
| 流程设计器 | React Flow | 11 |
| 图表 | ECharts | 5 |
| 鉴权 | JWT + Refresh Token | - |
| API 风格 | RESTful，前缀 `/api/v1/` | - |
| 容器化 | Docker + Docker Compose | - |
| CI/CD | GitHub Actions | - |
| 数据库迁移 | golang-migrate | - |

---

## 二、项目结构

```
StarByte/
├── backend/                          # Go 后端
│   ├── cmd/server/main.go            # 服务入口（优雅启停、路由注册）
│   ├── internal/                     # 业务模块（每个模块独立目录）
│   │   ├── user/                     # 用户模块（参考实现）
│   │   │   ├── handler/              #   接口层
│   │   │   ├── service/              #   业务层
│   │   │   ├── repo/                 #   数据访问层
│   │   │   ├── model/                #   数据模型
│   │   │   └── dto/                  #   请求/响应结构体
│   │   ├── audit/                    # 审计日志（空目录，待开发）
│   │   ├── auth/                     # 认证（空目录，待开发）
│   │   ├── file/                     # 文件管理（空目录，待开发）
│   │   ├── internship/               # 实习管理（空目录，待开发）
│   │   ├── interview/                # 面试管理（#7）
│   │   ├── meeting/                  # 会议管理 + 投票（#8）
│   │   ├── member/                   # 入会申请 + 人员档案（#6）
│   │   ├── notification/             # 通知系统（空目录，待开发）
│   │   ├── stats/                    # 数据统计（空目录，待开发）
│   │   ├── task/                     # 任务管理（空目录，待开发）
│   │   └── workflow/                 # 流程引擎（空目录，待开发）
│   ├── pkg/                          # 公共包
│   │   ├── config/                   #   配置管理（YAML）
│   │   ├── database/                 #   数据库连接（GORM）
│   │   ├── events/                   #   事件总线（发布/订阅）
│   │   ├── logger/                   #   日志（Zap 结构化日志）
│   │   ├── middleware/               #   中间件
│   │   │   ├── auth/                 #     JWT 鉴权 + 权限校验
│   │   │   └── common.go             #     RequestID/CORS/Logger/Recovery
│   │   ├── redis/                    #   Redis 客户端
│   │   ├── response/                 #   统一响应格式
│   │   └── utils/                    #   工具函数（密码哈希等）
│   ├── migrations/                   # 数据库迁移文件
│   │   ├── 000001_init_schema.*      #   用户/角色/权限/部门/审计/通知等
│   │   └── 000002_workflow_engine.*  #   流程引擎表
│   ├── configs/config.yaml           # 配置文件
│   ├── go.mod                        # Go 模块定义
│   └── Dockerfile
├── frontend/                         # React 前端
│   ├── src/
│   │   ├── api/                      # API 请求（request.ts 封装 Axios）
│   │   │   ├── request.ts            #   Axios 实例（自动 Token 刷新）
│   │   │   ├── auth.ts               #   认证 API
│   │   │   └── user.ts               #   用户 API
│   │   ├── components/               # 公共组件
│   │   ├── hooks/                    # 自定义 Hooks（usePermission 等）
│   │   ├── layouts/                  # 布局（MainLayout + Header）
│   │   ├── pages/                    # 页面
│   │   │   ├── login/                #   登录/注册页
│   │   │   ├── dashboard/            #   仪表盘（ECharts 示例）
│   │   │   └── user/                 #   用户列表（CRUD 示例）
│   │   ├── router/routes.tsx         # 路由配置（含全部菜单项）
│   │   ├── store/                    # Redux store
│   │   │   └── slices/               #   authSlice/userSlice/appSlice
│   │   ├── types/api.ts              # TypeScript 类型定义
│   │   ├── utils/storage.ts          # 本地存储工具
│   │   └── styles/global.css         # 全局样式
│   ├── Dockerfile + nginx.conf       # 前端容器化
│   └── package.json
├── deploy/docker-compose.yml        # Docker Compose 编排
├── docs/                             # 文档
│   ├── specs/                        #   设计文档
│   │   ├── 00-overall-architecture.md
│   │   ├── 01-workflow-engine.md
│   │   └── 02-rbac-system.md
│   └── dev-guide/                    #   开发规范
│       ├── ai-prompt-quick-start.md  #   AI 提示词快速开始
│       └── prompts/                  #   场景化 AI 提示词
│           ├── backend-prompt.md
│           ├── frontend-prompt.md
│           ├── fullstack-prompt.md
│           ├── review-prompt.md
│           └── workflow-prompt.md
├── .github/                          # GitHub 配置
│   ├── workflows/ci.yml              #   CI 流水线
│   ├── ISSUE_TEMPLATE/               #   Issue 模板
│   └── pull_request_template.md      #   PR 模板
├── CONTRIBUTING.md                   # 贡献指南
├── README.md
└── TEAM_DEV_GUIDE.md                 # ← 你正在看的这个文件
```

---

## 三、后端代码规范

### 3.1 模块四层架构（强制）

每个业务模块必须遵循以下结构，**依赖方向只能从上到下，禁止反向依赖**：

```
internal/{module}/
├── handler/        # 接口层：参数校验 → 调用 service → 返回响应
├── service/        # 业务层：核心逻辑、事务控制（先定义接口再实现）
├── repo/           # 数据层：GORM CRUD（不含业务逻辑）
├── model/          # 数据模型：GORM tag 完整，UUID 主键
└── dto/            # 请求/响应结构体（XxxRequest / XxxResponse）
```

```
handler → service → repo → model
  ↑ 唯一允许的依赖方向，禁止反向
```

**参考实现**：`backend/internal/user/` 是完整的四层实现，所有新模块请参照其结构。

### 3.2 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写简洁，无下划线 | `user`, `workflow` |
| 文件名 | snake_case | `user_service.go`, `flow_engine.go` |
| 结构体 | PascalCase | `UserService`, `FlowInstance` |
| 导出函数 | PascalCase | `CreateUser`, `GetByID` |
| 私有函数 | camelCase | `validateInput`, `buildQuery` |
| 常量 | PascalCase | `MaxPageSize`, `DefaultTimeout` |
| 错误码 | `ErrCodeXxx` | `ErrCodeUserNotFound` |
| Service 接口 | `XxxService` | `UserService` |
| Service 实现 | `xxxService`（小写） | `userService` |
| Repo 接口 | `XxxRepo` | `UserRepo` |
| 请求 DTO | `XxxRequest` | `CreateUserRequest` |
| 响应 DTO | `XxxResponse` | `UserListResponse` |

### 3.3 统一响应格式（强制）

所有 API 必须使用 `pkg/response` 包返回：

```go
// 成功（带数据）
response.OK(c, data)

// 成功（无数据）
response.OKWithoutData(c)

// 分页
response.Page(c, list, total, page, pageSize)

// 失败
response.Error(c, err)                           // 自动识别错误类型
response.BadRequest(c, "参数错误: xxx")           // 400
response.Unauthorized(c, "未授权")               // 401
response.Forbidden(c, "禁止访问")                // 403
response.NotFound(c, "资源不存在")               // 404
response.ErrorWithCode(c, 2002, "用户名或密码错误") // 自定义错误码
```

**JSON 响应结构**：
```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "request_id": "uuid-v4",
  "timestamp": 1692873600
}
```

**分页响应结构**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "page_size": 10
  },
  "request_id": "...",
  "timestamp": 1692873600
}
```

### 3.4 错误码分段（强制）

| 范围 | 模块 | 示例 |
|------|------|------|
| 0 | 成功 | - |
| 1000-1999 | 通用错误 | 1001 参数错误, 1002 未授权, 1003 禁止访问 |
| 2000-2999 | 用户模块 | 2001 用户已存在, 2002 用户名或密码错误 |
| 3000-3999 | 权限模块（RBAC） | 3001 角色不存在, 3002 权限不足 |
| 4000-4999 | 流程引擎 | 4001 流程定义不存在, 4002 流程已结束 |
| 5000-5999 | 审计日志 | 5001 审计日志不存在, 5002 导出格式不支持 |
| 6000-6999 | 会员模块 | 6001 申请不存在, 6003 重复申请, 6004 档案不存在 |
| 7000-7999 | 面试模块 | 7001 场次不存在, 7002 记录不存在, 7003 时间冲突 |
| 8000-8999 | 会议模块 | 8001 会议不存在, 8002 状态不允许, 8006 投票未开始或已结束 |
| 9000-9999 | 任务模块 | 9001 任务不存在, 9002 无权操作 |
| 10000-10999 | 实习模块 | 10001 实习不存在, 10002 无权操作 |
| 11000-11999 | 统计模块 | 11001 提供者不存在, 11002 参数无效 |
| 12000-12999 | 通知模块 | 12001 通知不存在, 12003 模板已存在 |

### 3.5 错误处理（强制）

```go
// ✅ 正确：包装错误，保留原始信息
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    if user == nil {
        return nil, response.NewError(2001, "用户不存在")
    }
    return user, nil
}

// ❌ 错误：忽略错误
user, _ := s.repo.GetByID(ctx, id)

// ❌ 错误：直接返回底层错误
return nil, err
```

### 3.6 GORM Model 规范

```go
type User struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
    Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
    PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // 敏感字段用 json:"-"
    Status       int            `gorm:"type:smallint;default:0;index" json:"status"`
    DepartmentID *uuid.UUID     `gorm:"type:uuid;index" json:"department_id"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
```

规则：
- 主键：UUID（`gorm:"type:uuid;primaryKey"`）
- 表名：复数形式（`users`, `roles`, `flow_definitions`）
- 软删除：`gorm.DeletedAt`
- 需要查询的字段加 `index`
- 敏感字段（密码等）用 `json:"-"`
- JSON 字段用 `gorm:"type:jsonb"`

### 3.7 Handler 层规范

```go
func (h *UserHandler) Login(c *gin.Context) {
    // 1. 参数绑定与校验
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }

    // 2. 调用 Service
    result, err := h.userService.Login(c.Request.Context(), &req, c.ClientIP())
    if err != nil {
        response.Error(c, err)
        return
    }

    // 3. 返回响应
    response.OK(c, result)
}
```

规则：
- Handler 只做三件事：参数校验、调用 Service、返回响应
- 禁止在 Handler 中写业务逻辑
- 路由注册函数：`RegisterXxxRoutes(r *gin.RouterGroup, handler *XxxHandler)`
- 每个接口写 Swagger 注释

### 3.8 Service 层规范

```go
// 1. 先定义接口
type UserService interface {
    Login(ctx context.Context, req *dto.LoginRequest, ip string) (*dto.LoginResponse, error)
    Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfoResponse, error)
}

// 2. 再实现
type userService struct {
    userRepo repo.UserRepo
    jwtConfig *config.JWTConfig
    redis    *redis.Client
}

func NewUserService(userRepo repo.UserRepo, jwtConfig *config.JWTConfig, redis *redis.Client) UserService {
    return &userService{userRepo: userRepo, jwtConfig: jwtConfig, redis: redis}
}
```

规则：
- 事务控制在 Service 层：`s.db.Transaction(func(tx *gorm.DB) error { ... })`
- 事件发布在 Service 层：通过 `pkg/events/event_bus.go`

### 3.9 数据库迁移规范

- 文件位置：`backend/migrations/`
- 命名：`{序号}_{描述}.up.sql` / `{序号}_{描述}.down.sql`
- 所有表结构变更必须通过迁移文件
- UUID 生成：`uuid_generate_v4()`
- 外键约束：`REFERENCES table_name(id)`
- 迁移文件一旦提交就不能修改，只能新增

### 3.10 日志规范

```go
import "github.com/Yogdunana/StarByte/backend/pkg/logger"

// 带上下文（包含 request_id）
logger.Ctx(ctx).Info("user login",
    zap.String("user_id", userID.String()),
    zap.String("username", username),
)
```

---

## 四、前端代码规范

### 4.1 目录与文件命名

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserList.tsx`, `Login.tsx` |
| Hook/工具文件 | camelCase | `usePermission.ts`, `storage.ts` |
| 组件名 | PascalCase | `UserList`, `MainLayout` |
| 函数/变量 | camelCase | `handleClick`, `isLoading` |
| 常量 | UPPER_SNAKE_CASE | `MAX_PAGE_SIZE` |
| 类型/接口 | PascalCase | `UserInfo`, `LoginRequest` |
| CSS Module 类名 | camelCase | `.container`, `.userCard` |
| Slice 名称 | camelCase | `authSlice`, `userSlice` |

### 4.2 组件规范

```tsx
// 使用函数组件 + Hooks
// Props 用 interface 定义
// 单个组件不超过 300 行
interface UserCardProps {
  user: UserInfo;
  onEdit?: (user: UserInfo) => void;
}

const UserCard: React.FC<UserCardProps> = ({ user, onEdit }) => {
  return (
    <Card title={user.real_name}>
      <p>{user.email}</p>
      {onEdit && <Button onClick={() => onEdit(user)}>编辑</Button>}
    </Card>
  );
};

export default UserCard;
```

### 4.3 TypeScript 规范

- 优先 `interface` 定义对象类型，`type` 定义联合类型
- **禁止 `any`**，万不得已用 `unknown`
- API 请求/响应类型必须完整定义

### 4.4 Redux Toolkit 规范

```typescript
// 按模块拆分 slice
// 异步 action 用 createAsyncThunk
// 同步 action 用 createSlice
// 提供 select 函数
export const fetchCurrentUser = createAsyncThunk(
  'user/fetchCurrentUser',
  async () => {
    return await getCurrentUser();
  }
);

export const selectCurrentUser = (state: RootState) => state.user.currentUser;
```

原则：
- 全局共享状态才放 Redux（用户信息、权限、主题）
- 组件内部状态用 `useState`
- 表单状态用 Ant Design `Form` 管理

### 4.5 API 请求规范

使用 `src/api/request.ts` 中封装的 Axios 实例（自动携带 Token、自动刷新）：

```typescript
// src/api/meeting.ts
import request from './request';
import type { Meeting } from '@/types/api';

export function getMeetingList(params: ListParams): Promise<PageResponse<Meeting>> {
  return request.get('/meetings', { params });
}

export function createMeeting(data: CreateMeetingRequest): Promise<Meeting> {
  return request.post('/meetings', data);
}
```

### 4.6 路由规范

```tsx
// src/router/routes.tsx
// 路由 meta 配置
{
  path: 'user/list',
  element: <UserList />,
  meta: {
    title: '用户列表',
    icon: 'UserOutlined',
    permission: 'user:read',  // 权限码
  },
}
```

### 4.7 权限控制

```tsx
// 前端只做 UI 控制，后端必须做权限校验
import { usePermission } from '@/hooks/usePermission';

const UserList: React.FC = () => {
  const canCreate = usePermission('user:create');
  return (
    <div>
      {canCreate && <Button>新增用户</Button>}
    </div>
  );
};
```

### 4.8 样式规范

- 优先使用 Ant Design 组件和主题
- 组件级样式用 CSS Modules（`.module.css`）
- 全局样式在 `styles/global.css`
- 最小支持 1280px 宽度

---

## 五、Git 工作流与提交规范

### 5.1 工作流：GitHub Flow + Squash Merge

```
main ──●───────────●───────────●─────  始终可部署
       │           │           │
       └── feature/xxx ──┘  feature/yyy ──┘
```

### 5.2 分支命名

| 前缀 | 用途 | 示例 |
|------|------|------|
| `feature/` | 新功能 | `feature/rbac-role-management` |
| `fix/` | Bug 修复 | `fix/login-token-expired` |
| `refactor/` | 重构 | `refactor/user-service` |
| `docs/` | 文档 | `docs/api-reference` |
| `chore/` | 构建/工具 | `chore/update-deps` |

命名使用 kebab-case（小写 + 连字符）。

### 5.3 Commit 规范：Conventional Commits

```
<type>(<scope>): <subject>

<body>

<footer>
```

| type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复 Bug |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响运行） |
| `refactor` | 重构 |
| `perf` | 性能优化 |
| `test` | 测试 |
| `chore` | 构建/工具 |
| `ci` | CI/CD 变更 |
| `revert` | 回滚 |

示例：
```
feat(workflow): 添加条件分支节点支持

- 实现排他网关节点类型
- 集成表达式引擎
- 支持配置分支条件

Closes #45
```

### 5.4 PR 规范

- 单个 PR 控制在 **500 行以内**（不含自动生成文件）
- 至少 **1 人 Code Review 通过**（核心模块需 2 人）
- CI 检查全部通过
- 使用 **Squash Merge**
- 前端变更必须附截图

### 5.5 提交前检查清单

```bash
# 后端
cd backend && go fmt ./... && go vet ./... && go test ./...

# 前端
cd frontend && npm run lint && npm run build
```

---

## 六、API 设计规范

### 6.1 路由前缀

所有 API 以 `/api/v1/` 为前缀。

### 6.2 RESTful 路由模板

```
# 标准 CRUD
GET    /api/v1/{module}            # 列表（分页+筛选）
POST   /api/v1/{module}            # 创建
GET    /api/v1/{module}/:id        # 详情
PUT    /api/v1/{module}/:id        # 更新
DELETE /api/v1/{module}/:id        # 删除

# 子资源
GET    /api/v1/{module}/:id/{sub}  # 子资源列表
POST   /api/v1/{module}/:id/{sub}  # 创建子资源

# 动作（非 CRUD）
POST   /api/v1/{module}/:id/{action}  # 执行动作
```

### 6.3 分页参数

```
GET /api/v1/users?page=1&page_size=10&keyword=张&status=0
```

| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量（最大 100） |
| keyword | string | - | 搜索关键词 |

### 6.4 鉴权

```
Authorization: Bearer <access_token>
```

- Access Token 过期后用 Refresh Token 刷新
- Token 刷新接口：`POST /api/v1/auth/refresh`

### 6.5 状态码

| HTTP 状态码 | 含义 |
|-------------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权（Token 无效/过期） |
| 403 | 禁止访问（权限不足） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 七、数据库规范

### 7.1 表设计原则

- 主键：UUID（`uuid_generate_v4()`）
- 表名：复数形式（`users`, `roles`, `meetings`）
- 时间字段：`created_at`, `updated_at`（`TIMESTAMP DEFAULT CURRENT_TIMESTAMP`）
- 软删除：`deleted_at TIMESTAMP`（可选）
- JSON 字段：`JSONB` 类型
- 状态字段：`SMALLINT`（0=正常，1=禁用，2=锁定 等）
- 外键约束：`REFERENCES table_name(id)`
- 需要查询的字段加索引

### 7.2 已有核心表

| 表名 | 说明 |
|------|------|
| users | 用户表 |
| roles | 角色表 |
| permissions | 权限表 |
| role_permissions | 角色-权限关联 |
| user_roles | 用户-角色关联 |
| departments | 部门表（树形） |
| positions | 职位表 |
| audit_logs | 审计日志 |
| notifications | 通知表 |
| notification_templates | 通知模板 |
| files | 文件表 |
| flow_definitions | 流程定义 |
| flow_definition_versions | 流程版本 |
| flow_instances | 流程实例 |
| flow_tasks | 流程任务 |
| flow_task_histories | 任务历史 |
| flow_variables | 流程变量 |
| member_applications | 入会申请 |
| member_profiles | 会员档案 |
| interviews | 面试记录 |
| interview_evaluations | 面试评分 |
| meetings | 会议表 |
| meeting_agendas | 会议议程 |
| meeting_votes | 投票表 |
| meeting_vote_records | 投票记录 |

---

## 八、AI 辅助开发指南

### 8.1 使用 TRAE 开发时的流程

> **⚠️ 强制规则：开始任何工作前，必须完成步骤 0-2，否则禁止编写代码。**

#### 步骤 0：检查优先级与认领资格

1. 打开 GitHub Issues 页面，按优先级标签筛选（`priority:p0` → `priority:p1` → `priority:p2` → `priority:p3`）
2. **若 P0 Issue 尚未全部完成（存在未关闭、或未标 `status:completed` 的 P0），禁止认领 P1/P2/P3 Issue**
3. **若 P1 核心 8 项完成不足 50%，禁止认领 P2 Issue**（口径见 §9.1，不要把 #47–#76 等扩展 P1 算进分母）
4. **若 P2 业务模块完成不足 80%，禁止认领 P3 Issue**
5. 每人**最多同时认领 2 个 Issue**，且**禁止同时认领 2 个 P0/P1 高优先级 Issue**

#### 步骤 1：按格式认领

在目标 Issue 评论区发布以下格式的认领评论：

```
我来认领 @Yogdunana
GitHub 账户名：[你的 GitHub 用户名]
预计完成时间：YYYY-MM-DD
关联分支：feature/[issue-number]-[brief-description]
```

等待项目负责人确认并分配（设置 Assignee + 添加 `status:claimed` 标签）后，方可开始开发。

#### 步骤 2：检查需求变更（每次开始编码前必做）

> **这是强制步骤。即使你之前已经认领，每次让 AI 写代码前都必须执行。**

1. 检查该 Issue 是否有 `req:changed` 标签
2. 检查 Issue 评论中是否有 `[REQ_CHANGE]` 开头的评论
3. 读取仓库根目录的 `REQUIREMENT_CHANGES.md`，查看与本 Issue 相关的未处理变更
4. 若存在未处理的变更：
   - 仔细阅读变更内容
   - 在 `[REQ_CHANGE]` 评论下回复：`已确认变更，开始适配。适配计划：[简述]`
   - 根据变更内容调整开发计划
5. 若无变更，正常继续开发

#### 步骤 3：开发

1. 从 main 拉取新分支：`git checkout -b feature/xxx`
2. 让 AI 读取本文档（将此文件链接发给 AI）
3. 描述你的具体需求 → AI 生成代码
4. 人工 review AI 生成的代码
5. 本地测试通过后提交
6. 提交 PR（使用 PR 模板）

### 8.2 给 AI 的系统提示词（可直接复制）

```
你是 StarByte 项目的高级开发工程师。请严格遵循以下项目规范进行开发。

项目仓库：https://github.com/Yogdunana/StarByte
请先阅读仓库中的 TEAM_DEV_GUIDE.md 文件，理解项目规范后再开始开发。

═════════════════════════════════════════
【强制】开发前检查清单（AI 必须按顺序执行）
═════════════════════════════════════════

1. 优先级检查：
   - 检查当前 Issue 的优先级标签（priority:p0 / p1 / p2 / p3）
   - 若存在未完成的 P0 Issue，且当前 Issue 不是 P0，提醒用户：「存在未完成的 P0 基础设施 Issue，建议优先认领 P0 任务」
   - P0 未完成时禁止开发 P1/P2/P3；P1 **核心 8 项**完成 <50% 时禁止开发 P2；P2 完成 <80% 时禁止开发 P3

2. 需求变更检查：
   - 使用 GitHub API 检查当前 Issue 是否有 req:changed 标签
   - 检查 Issue 评论中是否有 [REQ_CHANGE] 开头的评论
   - 读取仓库根目录 REQUIREMENT_CHANGES.md，查找与当前 Issue 相关的未处理变更
   - 若发现未处理的变更，停止开发，提醒用户：「检测到需求变更，请先阅读变更内容并回复确认后再继续」
   - 若无变更，继续开发

3. 认领确认检查：
   - 确认用户已在 Issue 评论区按格式认领（包含 GitHub 账户名、预计完成时间、关联分支）
   - 确认 Issue 已被分配（Assignee）且带有 status:claimed 标签
   - 若未认领，提醒用户先完成认领流程

═════════════════════════════════════════
关键规范摘要
═════════════════════════════════════════

1. 后端四层架构：handler → service → repo → model（禁止反向依赖）
2. 前端函数组件 + Hooks，TypeScript 严格模式
3. 主键 UUID，表名复数
4. 统一响应格式：{ code, message, data, request_id, timestamp }
5. 错误码按模块分段（用户 2000-2999，权限 3000-3999，流程 4000-4999，审计 5000-5999，会员 6000-6999，面试 7000-7999，会议 8000-8999，任务 9000-9999，实习 10000-10999，统计 11000-11999，通知 12000-12999）
6. Conventional Commits：feat/fix/docs/refactor/test/chore
7. PR 使用 Squash Merge，至少 1 人 Review
8. 单文件不超过 300 行
9. 提交前运行：后端 go fmt + go vet，前端 npm run lint
10. 权限校验在后端做，前端只做 UI 控制

═════════════════════════════════════════
认领格式（用户必须在 Issue 评论区发布）
═════════════════════════════════════════

我来认领 @Yogdunana
GitHub 账户名：[你的 GitHub 用户名]
预计完成时间：YYYY-MM-DD
关联分支：feature/[issue-number]-[brief-description]

═════════════════════════════════════════
参考实现
═════════════════════════════════════════

- 后端用户模块：backend/internal/user/（完整四层架构）
- 前端用户列表：frontend/src/pages/user/UserList.tsx
- 前端登录页：frontend/src/pages/login/Login.tsx
- API 封装：frontend/src/api/request.ts
```

### 8.3 注意事项

- AI 生成的代码**必须人工 review**，不能直接提交
- AI 可能产生幻觉（生成不存在的 API），必须核实
- 安全相关代码（鉴权、权限）要特别仔细检查
- 单个文件不超过 300 行，超过需拆分
- 遇到不确定的，查官方文档或问团队

---

## 九、任务认领与 Issue 列表

### 9.1 优先级体系

| 优先级 | 标签 | 说明 | 认领规则 |
|--------|------|------|----------|
| P0 | `priority:p0` | 核心基础设施，不完成则其他模块无法开发 | 必须最先完成，禁止跳过 |
| P1 | `priority:p1` | 核心功能模块，其他业务模块依赖 | P0 全部完成后才能认领 |
| P2 | `priority:p2` | 具体业务功能，模块间弱依赖 | P1 **核心 8 项**完成 50% 后可认领 |
| P3 | `priority:p3` | 辅助功能、优化、文档 | P2 业务模块完成 80% 后可认领 |

**P1 门禁口径（强制，2026-09-04 起）**：

- 「P1 核心」**只**统计 §9.4 的 8 个：#3 #4 #5 #17 #18 #19 #20 #21
- 2026-08-27 之后新增的 P1（#47–#50、#71–#76 等）是 **P1 扩展**，标题带 `[扩展]` / `[二期]` 的也算扩展
- 扩展 P1 **不计入** 50% 分母，也**不能**单独用来宣称「P1 已完成、可以认 P2」
- 完成判定：GitHub `state=closed` **且** 带 `status:completed`（已关闭但仍标 `status:available` / `status:claimed` 的，一律不算完成）

当前口径快照（2026-09-04）：P1 核心 **8/8 已完成**，**可以认领 P2**。

**认领限制**：
- 每人**最多同时认领 2 个 Issue**
- **禁止同时认领 2 个 P0/P1 高优先级 Issue**（可 1 个 P1 + 1 个 P2，或 1 个 P2 + 1 个 P3）
- 若需放弃认领，必须在评论区说明并通知 @Yogdunana 移除 `status:claimed` 标签，恢复 `status:available`

### 9.2 认领格式（强制）

**必须在 Issue 评论区按以下格式认领**：

```
我来认领 @Yogdunana
GitHub 账户名：[你的 GitHub 用户名]
预计完成时间：YYYY-MM-DD
关联分支：feature/[issue-number]-[brief-description]
```

项目负责人确认后，会执行以下操作：
1. 将你设为 Issue 的 Assignee
2. 添加 `status:claimed` 标签
3. 移除 `status:available` 标签

**三重防撞机制**：评论区认领记录 + GitHub Assignees + `status:claimed` 标签，三者缺一不可。缺任一项则认领无效，Issue 保持 `status:available`。

**完成后（强制）**：Squash Merge 并关闭 Issue 后，必须：
1. 去掉 `status:available` / `status:claimed` / `status:in-progress` / `status:review`
2. 加上 `status:completed`
3. 确认 Assignee 为实际完成者（不要留空）
4. PR 正文不要对「只完成一部分」的 Issue 写 `Closes #N`（会提前关单；用 `Related to #N`）

### 9.3 需求变更机制

> **目的**：确保开发者认领任务后，若需求发生变更，开发者和 AI 都能及时感知并适配。

#### 9.3.1 变更发布流程（项目负责人操作）

当需求需要变更时，项目负责人执行以下操作：

1. 在相关 Issue 评论区发布变更通知：
   ```
   [REQ_CHANGE]
   变更类型：[新增/修改/删除]
   影响范围：[API/数据模型/UI/流程/权限]
   具体内容：[变更详情]
   是否需要返工：[是/否，说明范围]
   变更编号：RC-YYYYMMDD-NNN
   ```

2. 在该 Issue 上添加 `req:changed` 标签

3. 更新仓库根目录 `REQUIREMENT_CHANGES.md`，添加变更记录

#### 9.3.2 开发者响应流程（强制）

开发者在每次让 AI 写代码前，必须执行以下检查：

1. **检查标签**：查看 Issue 是否有 `req:changed` 标签
2. **检查评论**：查看 Issue 评论中是否有 `[REQ_CHANGE]` 开头的评论
3. **检查日志**：读取仓库根目录 `REQUIREMENT_CHANGES.md`，查找与当前 Issue 相关的未处理变更
4. **响应变更**：若发现未处理的变更，在 `[REQ_CHANGE]` 评论下回复：
   ```
   已确认变更，开始适配。适配计划：[简述如何调整]
   ```
5. **完成适配后**：在 `[REQ_CHANGE]` 评论下回复：
   ```
   变更适配完成。关联 PR：#XXX
   ```
   项目负责人确认后将 `REQUIREMENT_CHANGES.md` 中对应条目标记为「已完成」

#### 9.3.3 AI 开发提示词中的变更检查指令

AI 系统提示词（见 8.2 节）已包含强制变更检查步骤。当开发者告诉 AI「我要开发 Issue #X」时，AI 会自动：
1. 通过 GitHub API 检查 Issue 标签和评论
2. 读取 `REQUIREMENT_CHANGES.md`
3. 若发现未处理变更，停止开发并提醒开发者先处理变更

### 9.4 一期核心 Issue 列表（33 个，门禁用这一张表）

> GitHub 上还有 #46–#102 等扩展单。**认领门禁、完成度统计只看本表。**  
> 状态以 GitHub 为准；下表「完成」表示已 `status:completed`（2026-09-04 治理后）。

#### P0 - 核心基础设施（7 个，必须最先完成）— 已全部完成

| Issue | 标题 | 模块 | 依赖 | 状态 |
|-------|------|------|------|------|
| [#1](https://github.com/Yogdunana/StarByte/issues/1) | RBAC 权限系统 + 组织架构 | backend | 无 | 完成 |
| [#2](https://github.com/Yogdunana/StarByte/issues/2) | 流程引擎核心模块 | backend | 无 | 完成（#38–#43；勿认 #27 重做网关） |
| #12 | 后端错误处理与统一异常管理 | backend | 无 | 完成 |
| #13 | 配置中心与多环境管理（YAML/环境变量） | backend | 无 | 完成（运行时 configs 表见扩展 #47） |
| #14 | API 网关与路由中间件 | backend | 无 | 完成 |
| #15 | 前端公共组件库与布局系统 | frontend | 无 | 完成 |
| #16 | Docker 容器化与 CI/CD 完善 | devops | 无 | 完成 |

#### P1 - 核心服务（8 个，依赖 P0）— 门禁分母，8/8 已完成

| Issue | 标题 | 模块 | 依赖 | 状态 |
|-------|------|------|------|------|
| [#3](https://github.com/Yogdunana/StarByte/issues/3) | 前端流程设计器（React Flow） | frontend | #2 | 完成 |
| [#4](https://github.com/Yogdunana/StarByte/issues/4) | 消息通知系统 | fullstack | #1 | 完成（模板已含；勿认 #49） |
| [#5](https://github.com/Yogdunana/StarByte/issues/5) | 审计日志系统 | backend | #1 | 完成（二期 Diff/报告见 #76） |
| #17 | 用户认证服务（JWT + Refresh Token + 第三方登录预留） | backend | #1 | 完成（OAuth 落地见 #63） |
| #18 | 文件上传与管理服务（MinIO 集成） | backend | #1 | 完成（#111；须跑 `000009_file_management`） |
| #19 | 数据库迁移与种子数据 | backend | #1 | 完成（#112；业务表从 `000010` 起编） |
| #20 | 前端权限路由与菜单系统 | frontend | #1, #15 | 完成 |
| #21 | 前端 API 层封装与类型定义 | frontend | #15 | 完成 |

#### P2 - 业务模块（13 个，依赖 P1）

| Issue | 标题 | 模块 | 依赖 | 备注 |
|-------|------|------|------|------|
| [#6](https://github.com/Yogdunana/StarByte/issues/6) | 入会申请 + 人员档案 | fullstack | #1, #2, #4 | 完成（`000020` 补列；勿另起表） |
| [#7](https://github.com/Yogdunana/StarByte/issues/7) | 面试管理 | fullstack | #1, #2, #6 | 完成（`000021` 补列；勿另起表） |
| [#8](https://github.com/Yogdunana/StarByte/issues/8) | 会议管理 + 投票系统（等权 + 加权） | fullstack | #1, #4 | 进行中（`000022` 补列；勿另起表） |
| [#9](https://github.com/Yogdunana/StarByte/issues/9) | 任务流转 | fullstack | #1, #4 | |
| [#10](https://github.com/Yogdunana/StarByte/issues/10) | IT 实习管理 | fullstack | #1, #4 | |
| [#11](https://github.com/Yogdunana/StarByte/issues/11) | 数据统计与可视化报表 | fullstack | #1, #5 | |
| #22 | 财务管理模块（一期预留接口） | fullstack | #1, #4 | |
| #23 | 纪律处分记录模块 | fullstack | #1, #4 | |
| #24 | 合同管理模块（一期预留接口） | fullstack | #1, #4 | |
| #25 | 邮件通知**增强**（附件/批量/重试/记录） | backend | #4 | 2026-09-04 改写；勿重做 SMTP |
| #26 | 前端仪表盘与数据大屏 | frontend | #11, #15 | |
| #27 | ~~流程引擎节点插件~~ | backend | #2 | **已关闭（重复 #2）** |
| #28 | 前端表单引擎与动态表单 | frontend | #3, #15 | |

#### 易撞车的扩展 Issue（不在 33 张门禁表内）

| Issue | 怎么处理 |
|-------|----------|
| #47 | 运行时 `configs` 表，**不是** #13 |
| #49 | **已关闭**，模板已在 #4 |
| #51 | **已关闭**，加权投票在 #8 |
| #63 | #17 预留接口落地，禁止重写认证 |
| #76 | #5 二期（Diff/报告），错误码仍 5000–5999 |
| #99 | 仅会签/加签/委派/超时/包容网关；网关本体已在 #2 |

#### P3 - 支持功能（5 个）

| Issue | 标题 | 模块 | 依赖 |
|-------|------|------|------|
| #29 | API 文档生成与 Swagger 集成 | backend | #14 |
| #30 | 单元测试与集成测试框架 | fullstack | P0/P1 完成 |
| #31 | 前端国际化与主题切换 | frontend | #15 |
| #32 | 性能监控与健康检查 | backend | #14 |
| #33 | 开发者文档与 README 完善 | docs | 各模块基本完成 |

### 9.5 标签说明

| 标签 | 含义 |
|------|------|
| `status:available` | 可认领（仅开放 Issue） |
| `status:claimed` | 已认领（仅开放 Issue；完成后必须去掉） |
| `status:in-progress` | 开发中 |
| `status:blocked` | 阻塞中 |
| `status:review` | 待 Review |
| `status:completed` | 已完成并关闭（关闭后必须打上，同时去掉 available/claimed） |
| `priority:p0` | P0 - 核心基础设施（阻塞） |
| `priority:p1` | P1 - 核心功能模块 |
| `priority:p2` | P2 - 业务功能 |
| `priority:p3` | P3 - 支持功能 |
| `req:changed` | 需求已变更，开发者需检查 |
| `module:backend` | 纯后端 |
| `module:frontend` | 纯前端 |
| `module:fullstack` | 前后端 |
| `module:devops` | 运维/基础设施 |
| `module:docs` | 文档 |

---

## 十、本地开发环境

### 快速启动

```bash
# 1. 克隆仓库
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte

# 2. 启动基础设施（PostgreSQL + Redis + MinIO）
docker-compose -f deploy/docker-compose.yml up -d postgres redis minio

# 3. 后端
cd backend
cp .env.example .env
go mod download
go run cmd/server/main.go

# 4. 前端（新终端）
cd frontend
npm install
npm run dev
```

### 访问地址

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:5173 |
| 后端 API | http://localhost:8080/api/v1 |
| MinIO 控制台 | http://localhost:9001 (minioadmin/minioadmin) |

### 数据库迁移

```bash
# 安装 migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 运行迁移
migrate -path backend/migrations \
  -database "postgres://starbyte:starbyte@localhost:5432/starbyte?sslmode=disable" \
  up
```

---

## 十一、其他文档索引

| 文档 | 路径 | 说明 |
|------|------|------|
| 整体架构设计 | `docs/specs/00-overall-architecture.md` | 系统架构、模块划分 |
| 工作流引擎设计 | `docs/specs/01-workflow-engine.md` | 流程引擎核心设计 |
| RBAC 权限设计 | `docs/specs/02-rbac-system.md` | 权限系统设计 |
| 后端开发规范 | `docs/dev-guide/backend.md` | 详细后端规范 |
| 前端开发规范 | `docs/dev-guide/frontend.md` | 详细前端规范 |
| Git 工作流 | `docs/dev-guide/git-workflow.md` | 分支管理与 PR |
| PR 规范 | `docs/dev-guide/pr-specification.md` | PR 模板与审查 |
| AI 提示词 | `docs/dev-guide/prompts/` | 各场景 AI 提示词 |
| 贡献指南 | `CONTRIBUTING.md` | 完整贡献流程 |
