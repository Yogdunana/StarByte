# 后端开发规范

> **版本**: v1.0
> **日期**: 2026-08-24
> **适用范围**: StarByte 后端 Go 项目

---

## 1. 项目结构规范

### 1.1 模块目录结构

每个业务模块必须遵循以下四层结构：

```
module_name/
├── handler/        # 接口层
│   └── xxx_handler.go
├── service/        # 业务逻辑层
│   └── xxx_service.go
├── repo/           # 数据访问层
│   └── xxx_repo.go
├── model/          # 数据模型
│   └── xxx.go
├── dto/            # 数据传输对象
│   ├── request.go
│   └── response.go
└── module.go       # 模块入口（注册路由、依赖注入）
```

### 1.2 依赖方向

```
handler → service → repo → model
```

- handler 只能调用 service，不能直接调用 repo
- service 可以调用 repo 和其他 service
- repo 只能操作数据库，不能调用 service
- 禁止反向依赖

### 1.3 公共包结构

```
pkg/
├── config/         # 配置管理
├── database/       # 数据库连接与初始化
├── redis/          # Redis 客户端
├── minio/          # MinIO 客户端
├── middleware/     # Gin 中间件
│   ├── auth.go     # JWT 鉴权
│   ├── audit.go    # 审计日志
│   ├── cors.go     # 跨域
│   ├── limiter.go  # 限流
│   └── recovery.go # 恐慌恢复
├── response/       # 统一响应格式
├── logger/         # 日志封装
├── utils/          # 工具函数
├── ws/             # WebSocket 管理
└── events/         # 事件总线
```

---

## 2. 命名规范

### 2.1 包命名
- 使用小写、简洁的名字
- 不使用下划线、混合大小写
- 包名与目录名保持一致

### 2.2 文件命名
- 小写 + 下划线（snake_case）
- 测试文件：`xxx_test.go`
- 示例：`user_handler.go`、`member_service.go`

### 2.3 结构体命名
- 大驼峰（PascalCase）
- 请求 DTO：`XxxRequest`
- 响应 DTO：`XxxResponse`
- Service 接口：`XxxService`
- Service 实现：`xxxService`（小写开头，对外通过接口暴露）
- Repository 接口：`XxxRepo`
- Model：和表名对应，大驼峰

### 2.4 函数/方法命名
- 大驼峰（导出）或小驼峰（私有）
- CRUD 方法命名：
  - 创建：`Create` / `Add`
  - 查询单个：`GetByID` / `GetByXxx`
  - 查询列表：`List` / `Query`
  - 更新：`Update`
  - 删除：`Delete` / `Remove`

### 2.5 常量命名
- 大驼峰（导出）或小驼峰（私有）
- 相关常量用分组
- 错误码常量：`ErrCodeXxx`

---

## 3. 代码风格

### 3.1 基本规范
- 使用 `gofmt` 格式化代码
- 使用 `goimports` 自动管理 import
- import 分组：标准库 → 第三方库 → 内部包，每组之间空行
- 每行不超过 120 字符
- 函数不超过 50 行，超过考虑拆分

### 3.2 错误处理
- 永远不要忽略错误，使用 `_` 显式忽略并注释原因
- 错误使用 `fmt.Errorf("xxx: %w", err)` 包装，保留原始错误
- 自定义错误使用 `pkg/response` 中的错误类型
- Service 层返回原始 error，Handler 层统一转换为响应

```go
// 正确示例
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    if user == nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}
```

### 3.3 错误码规范

| 错误码范围 | 模块 |
|-----------|------|
| 0 | 成功 |
| 1000-1999 | 通用错误 |
| 2000-2999 | 用户模块 |
| 3000-3999 | 权限模块 |
| 4000-4999 | 流程引擎 |
| 5000-5999 | 会员模块 |
| 6000-6999 | 面试模块 |
| 7000-7999 | 会议模块 |
| 8000-8999 | 任务模块 |
| 9000-9999 | 通知模块 |

---

## 4. Handler 层规范

### 4.1 职责
- 接收 HTTP 请求
- 参数绑定与校验（使用 binding tag）
- 调用 Service 层
- 统一响应格式
- 不包含业务逻辑

### 4.2 请求参数校验
- 使用 `binding` tag 进行参数校验
- 路径参数：`uri` tag
- Query 参数：`form` tag
- Body 参数：`json` tag

```go
type ListUserRequest struct {
    Page     int    `form:"page,default=1" binding:"min=1"`
    PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
    Keyword  string `form:"keyword"`
    Status   *int   `form:"status"`
}
```

### 4.3 响应格式
- 统一使用 `pkg/response` 包
- 成功：`response.OK(c, data)`
- 失败：`response.Error(c, err)` / `response.ErrorWithCode(c, code, msg)`
- 分页：`response.Page(c, list, total, page, pageSize)`

### 4.4 路由注册
- 路由分组按模块划分
- 需要鉴权的路由放在 `auth` 组内
- 需要特定权限的路由使用权限中间件

```go
func RegisterRoutes(r *gin.RouterGroup, handler *UserHandler) {
    users := r.Group("/users")
    {
        users.GET("", handler.List)
        users.GET("/:id", handler.GetByID)
        users.POST("", handler.Create)
        users.PUT("/:id", handler.Update)
        users.DELETE("/:id", handler.Delete)
    }
}
```

---

## 5. Service 层规范

### 5.1 职责
- 核心业务逻辑
- 事务控制
- 调用 Repository 层
- 调用其他 Service
- 事件发布

### 5.2 事务管理
- 使用 GORM 事务
- Service 层方法中开启事务
- 多个 Repository 操作需要事务包裹

```go
func (s *userService) Create(ctx context.Context, req *CreateUserRequest) (*User, error) {
    err := s.db.Transaction(func(tx *gorm.DB) error {
        // 操作1
        if err := s.userRepo.Create(ctx, tx, user); err != nil {
            return err
        }
        // 操作2
        if err := s.userRoleRepo.BatchCreate(ctx, tx, userRoles); err != nil {
            return err
        }
        return nil
    })
    return user, err
}
```

### 5.3 接口定义
- Service 先定义接口，再实现
- 通过接口依赖，便于单元测试和 mock

```go
type UserService interface {
    Create(ctx context.Context, req *CreateUserRequest) (*User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    List(ctx context.Context, req *ListUserRequest) ([]User, int64, error)
    Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

---

## 6. Repository 层规范

### 6.1 职责
- 数据库 CRUD 操作
- 不包含业务逻辑
- 只做数据访问

### 6.2 GORM Model 定义
- 主键使用 UUID
- 使用 `gorm.Model` 或自定义基础模型
- 表名使用复数形式
- 字段注释要完整

```go
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
    PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
    RealName     string    `gorm:"type:varchar(50)" json:"real_name"`
    Status       int       `gorm:"type:smallint;default:0;index" json:"status"`
    DepartmentID *uuid.UUID `gorm:"type:uuid;index" json:"department_id"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    *time.Time `gorm:"index" json:"-"`
}

func (User) TableName() string {
    return "users"
}
```

### 6.3 查询规范
- 使用 `context.Context` 传递上下文
- 分页查询使用 `Limit` + `Offset`
- 条件查询链式调用
- 避免 N+1 查询，合理使用 Preload

---

## 7. DTO 规范

### 7.1 分类
- 请求 DTO：放在 `dto/request.go` 中
- 响应 DTO：放在 `dto/response.go` 中

### 7.2 命名
- 请求：`CreateXxxRequest`、`UpdateXxxRequest`、`ListXxxRequest`
- 响应：`XxxResponse`、`XxxDetailResponse`

### 7.3 注意事项
- 不要把 Model 直接作为响应返回
- 敏感字段（如密码）不能出现在响应中
- 时间字段统一使用 RFC3339 格式

---

## 8. 日志规范

### 8.1 日志级别
- `Debug`：调试信息，开发环境使用
- `Info`：正常业务流程信息
- `Warn`：警告信息，不影响主流程
- `Error`：错误信息，需要关注和处理

### 8.2 日志字段
- 必须包含：request_id、user_id（如果有）
- 错误日志包含：err、stack（可选）
- 使用结构化日志（zap）

```go
logger.Ctx(ctx).Info("user login",
    zap.String("user_id", userID.String()),
    zap.String("username", username),
)
```

### 8.3 上下文传递
- 使用 `context.Context` 传递 request_id 等上下文信息
- 通过 `logger.Ctx(ctx)` 获取带上下文的 logger

---

## 9. 配置规范

### 9.1 配置文件
- 使用 YAML 格式
- 配置文件放在 `configs/` 目录
- 支持环境变量覆盖

### 9.2 配置结构
```yaml
server:
  port: 8080
  mode: debug  # debug/release

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: starbyte
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key
  access_token_ttl: 900        # 15分钟（秒）
  refresh_token_ttl: 604800     # 7天（秒）

minio:
  endpoint: localhost:9000
  access_key: minioadmin
  secret_key: minioadmin
  use_ssl: false
  bucket: starbyte

email:
  enabled: true
  host: smtp.example.com
  port: 587
  username: noreply@example.com
  password: password
  from: "StarByte <noreply@example.com>"
```

---

## 10. 测试规范

### 10.1 单元测试
- 测试文件和源文件同目录，命名 `xxx_test.go`
- 测试覆盖率目标：核心模块 ≥ 70%
- 使用 testify 断言库
- 使用 mock 隔离外部依赖

### 10.2 测试命令
```bash
# 运行所有测试
go test ./...

# 运行指定模块测试
go test ./internal/user/... -v

# 查看覆盖率
go test ./... -cover
```

---

## 11. 数据库迁移规范

### 11.1 迁移文件命名
```
{序号}_{描述}.up.sql
{序号}_{描述}.down.sql
```

示例：
```
000001_init_schema.up.sql
000001_init_schema.down.sql
000002_add_user_identities.up.sql
000002_add_user_identities.down.sql
```

### 11.2 迁移命令
```bash
# 升级到最新版本
migrate -path ./migrations -database "postgres://..." up

# 降级 1 个版本
migrate -path ./migrations -database "postgres://..." down 1

# 强制指定版本（修复脏迁移）
migrate -path ./migrations -database "postgres://..." force 000001
```

### 11.3 注意事项
- 所有表结构变更都必须通过迁移文件，不能手动改表
- 迁移文件一旦提交就不能修改，只能新增
- up.sql 和 down.sql 必须对应，down 要能正确回滚
