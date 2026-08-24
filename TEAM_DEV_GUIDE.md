# StarByte 团队开发规范总纲

> **本文档是 StarByte 项目唯一的一站式开发参考。**
> **所有团队成员（含 AI 助手）在开始任何开发工作前，必须先阅读本文档。**
>
> 仓库地址：https://github.com/Yogdunana/StarByte
> 最后更新：2026-08-24

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
│   │   ├── interview/                # 面试管理（空目录，待开发）
│   │   ├── meeting/                  # 会议管理（空目录，待开发）
│   │   ├── member/                   # 会员管理（空目录，待开发）
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
| 6000-6999 | 会员模块 | 6001 申请不存在, 6002 状态不允许操作 |
| 7000-7999 | 面试模块 | 7001 面试场次不存在, 7002 时间冲突 |
| 8000-8999 | 会议模块 | 8001 会议不存在, 8002 投票已结束 |
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

1. 在 GitHub Issues 中找到你要开发的任务
2. 评论 `我来认领` 领取任务
3. 从 main 拉取新分支：`git checkout -b feature/xxx`
4. 让 AI 读取本文档（将此文件链接发给 AI）
5. 描述你的具体需求 → AI 生成代码
6. 人工 review AI 生成的代码
7. 本地测试通过后提交
8. 提交 PR（使用 PR 模板）

### 8.2 给 AI 的系统提示词（可直接复制）

```
你是 StarByte 项目的高级开发工程师。请严格遵循以下项目规范进行开发。

项目仓库：https://github.com/Yogdunana/StarByte
请先阅读仓库中的 TEAM_DEV_GUIDE.md 文件，理解项目规范后再开始开发。

关键规范摘要：
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

参考实现：
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

### Phase 1 - 核心基础（里程碑 #1，截止 9/10）

| Issue | 标题 | 优先级 | 模块 |
|-------|------|--------|------|
| [#1](https://github.com/Yogdunana/StarByte/issues/1) | RBAC 权限系统 + 组织架构 | critical | backend |
| [#2](https://github.com/Yogdunana/StarByte/issues/2) | 流程引擎核心模块 | critical | backend |
| [#3](https://github.com/Yogdunana/StarByte/issues/3) | 前端流程设计器（React Flow） | high | frontend |
| [#4](https://github.com/Yogdunana/StarByte/issues/4) | 消息通知系统 | high | fullstack |
| [#5](https://github.com/Yogdunana/StarByte/issues/5) | 审计日志系统 | high | backend |

### Phase 1 - 业务模块（里程碑 #2，截止 9/15）

| Issue | 标题 | 优先级 | 模块 |
|-------|------|--------|------|
| [#6](https://github.com/Yogdunana/StarByte/issues/6) | 入会申请 + 人员档案 | high | fullstack |
| [#7](https://github.com/Yogdunana/StarByte/issues/7) | 面试管理 | high | fullstack |
| [#8](https://github.com/Yogdunana/StarByte/issues/8) | 会议管理 + 投票系统 | high | fullstack |
| [#9](https://github.com/Yogdunana/StarByte/issues/9) | 任务流转 | medium | fullstack |
| [#10](https://github.com/Yogdunana/StarByte/issues/10) | IT 实习管理 | medium | fullstack |
| [#11](https://github.com/Yogdunana/StarByte/issues/11) | 数据统计与可视化报表 | medium | fullstack |

### 认领方式

在对应 Issue 下评论 `我来认领` 并 @Yogdunana，会将 Issue 分配给你。

### 标签说明

| 标签 | 含义 |
|------|------|
| `status:available` | 可认领 |
| `status:claimed` | 已认领 |
| `status:in-progress` | 开发中 |
| `status:blocked` | 阻塞中 |
| `priority:critical` | 紧急（其他模块依赖） |
| `priority:high` | 高优先级 |
| `priority:medium` | 中优先级 |
| `module:backend` | 纯后端 |
| `module:frontend` | 纯前端 |
| `module:fullstack` | 前后端 |

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