# RBAC 权限系统设计文档

> **版本**: v1.0
> **日期**: 2026-08-24
> **模块**: rbac
> **状态**: 一期设计

---

## 1. 设计目标

- 基于 RBAC（Role-Based Access Control）模型，支持细粒度权限控制
- 支持 API 级权限 + 数据级权限（部门范围）
- 完全配置化，所有权限和角色均可在后台管理页面动态配置，不硬编码
- 支持权限开关，社长可随时启用/禁用某个功能模块的权限
- 支持多角色，一个用户可以拥有多个角色
- 预留扩展能力，支持更复杂的权限模型（如 ABAC）

---

## 2. 核心概念

### 2.1 用户 (User)
系统的使用者，UUID 主键。一个用户可以拥有多个角色。

### 2.2 角色 (Role)
权限的集合。角色有层级（可选），支持角色继承。

### 2.3 权限 (Permission)
系统中的一个操作权限，如"创建用户"、"删除会员"、"查看报表"等。

### 2.4 资源 (Resource)
权限作用的对象，如"用户"、"会员"、"面试"等。

### 2.5 操作 (Action)
对资源的操作，如"创建"、"读取"、"更新"、"删除"、"审批"等。

### 2.6 数据权限 (Data Scope)
用户能看到的数据范围：
- `all`：全部数据
- `department`：本部门数据
- `department_and_sub`：本部门及下级部门
- `self`：仅自己的数据
- `custom`：自定义（指定部门）

---

## 3. 数据模型

### 3.1 users（用户表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `username` | VARCHAR(50) | 用户名（登录用，唯一） |
| `password_hash` | VARCHAR(255) | 密码哈希（bcrypt） |
| `real_name` | VARCHAR(50) | 真实姓名 |
| `avatar_url` | VARCHAR(500) | 头像URL |
| `email` | VARCHAR(100) | 邮箱 |
| `phone` | VARCHAR(20) | 手机号 |
| `gender` | SMALLINT | 性别：0=未知 1=男 2=女 |
| `status` | SMALLINT | 状态：0=正常 1=禁用 2=锁定 |
| `department_id` | UUID | 所属部门ID |
| `position_id` | UUID | 职务ID |
| `last_login_at` | TIMESTAMP | 最后登录时间 |
| `last_login_ip` | VARCHAR(50) | 最后登录IP |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |
| `deleted_at` | TIMESTAMP | 删除时间（软删除） |

### 3.2 user_identities（用户身份关联表）

用于支持一人多学号（本科/研究生同一人）。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `user_id` | UUID | 用户ID |
| `identity_type` | VARCHAR(20) | 身份类型：student_id(学号) / employee_id(工号) / email / phone |
| `identity_value` | VARCHAR(100) | 身份值（具体的学号/工号） |
| `is_primary` | BOOLEAN | 是否主身份 |
| `created_at` | TIMESTAMP | 创建时间 |

### 3.3 roles（角色表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `name` | VARCHAR(50) | 角色名称 |
| `code` | VARCHAR(50) | 角色编码（唯一，用于代码中引用） |
| `description` | VARCHAR(255) | 角色描述 |
| `parent_id` | UUID | 父角色ID（用于角色继承） |
| `sort_order` | INT | 排序 |
| `status` | SMALLINT | 状态：0=启用 1=禁用 |
| `is_system` | BOOLEAN | 是否系统内置角色（不可删除） |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

### 3.4 permissions（权限表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `name` | VARCHAR(100) | 权限名称 |
| `code` | VARCHAR(100) | 权限编码（唯一，如 `user:create`） |
| `resource` | VARCHAR(50) | 资源类型（如 user, member, interview） |
| `action` | VARCHAR(50) | 操作类型（如 create, read, update, delete） |
| `description` | VARCHAR(255) | 权限描述 |
| `parent_id` | UUID | 父权限ID（用于树形结构） |
| `sort_order` | INT | 排序 |
| `type` | SMALLINT | 类型：1=菜单 2=按钮 3=接口 |
| `path` | VARCHAR(255) | 前端路由路径（type=菜单时） |
| `icon` | VARCHAR(50) | 菜单图标（type=菜单时） |
| `api_method` | VARCHAR(10) | API 请求方法（type=接口时） |
| `api_path` | VARCHAR(255) | API 路径（type=接口时） |
| `is_system` | BOOLEAN | 是否系统内置权限（不可删除） |
| `status` | SMALLINT | 状态：0=启用 1=禁用 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

### 3.5 role_permissions（角色-权限关联表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `role_id` | UUID | 角色ID |
| `permission_id` | UUID | 权限ID |
| `data_scope` | VARCHAR(20) | 数据权限范围：all/department/department_and_sub/self/custom |
| `created_at` | TIMESTAMP | 创建时间 |

**唯一约束**：(role_id, permission_id)

### 3.6 user_roles（用户-角色关联表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `user_id` | UUID | 用户ID |
| `role_id` | UUID | 角色ID |
| `expired_at` | TIMESTAMP | 过期时间（NULL=永久） |
| `created_at` | TIMESTAMP | 创建时间 |

**唯一约束**：(user_id, role_id)

### 3.7 role_data_scopes（角色数据权限-自定义部门）

当 `data_scope = custom` 时，关联的部门列表。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `role_permission_id` | UUID | 角色权限关联ID |
| `department_id` | UUID | 部门ID |
| `created_at` | TIMESTAMP | 创建时间 |

---

## 4. 组织架构数据模型

### 4.1 departments（部门表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `name` | VARCHAR(100) | 部门名称 |
| `code` | VARCHAR(50) | 部门编码（唯一） |
| `parent_id` | UUID | 父部门ID |
| `leader_id` | UUID | 部门负责人ID |
| `description` | VARCHAR(255) | 部门描述 |
| `sort_order` | INT | 排序 |
| `status` | SMALLINT | 状态：0=启用 1=禁用 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

### 4.2 positions（职务表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `name` | VARCHAR(50) | 职务名称（如 会长、副会长、部长、副部长、干事） |
| `code` | VARCHAR(50) | 职务编码（唯一） |
| `level` | INT | 职务级别（数字越大级别越高） |
| `vote_weight` | DECIMAL(5,2) | 投票权重（用于加权投票） |
| `description` | VARCHAR(255) | 职务描述 |
| `sort_order` | INT | 排序 |
| `status` | SMALLINT | 状态：0=启用 1=禁用 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

---

## 5. 权限校验机制

### 5.1 后端权限校验

**中间件方式**：Gin 中间件自动校验 API 权限。

```go
// PermissionMiddleware 权限校验中间件
func PermissionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 JWT 中获取用户ID
        userID := GetUserID(c)
        
        // 2. 获取请求的 API 路径和方法
        apiPath := c.FullPath()
        method := c.Request.Method
        
        // 3. 查询用户是否拥有该 API 的权限
        hasPermission := rbacService.CheckAPIPermission(userID, apiPath, method)
        if !hasPermission {
            response.Forbidden(c, "无权限访问")
            c.Abort()
            return
        }
        
        // 4. 权限校验通过，将用户权限信息注入上下文
        c.Set("userPermissions", userPermissions)
        c.Next()
    }
}
```

### 5.2 数据权限过滤

在 Service 层或 Repo 层进行数据范围过滤。

```go
// GetDataScopeCondition 获取数据权限的 SQL 条件
func GetDataScopeCondition(userID uuid.UUID, resource string) (string, []interface{}) {
    // 查询用户对该资源的数据权限范围
    scope := rbacService.GetDataScope(userID, resource)
    
    switch scope {
    case "all":
        return "", nil  // 不过滤
    case "department":
        return "department_id = ?", []interface{}{userDeptID}
    case "department_and_sub":
        // 查询所有子部门ID
        deptIDs := getDepartmentAndSubIDs(userDeptID)
        return "department_id IN (?)", []interface{}{deptIDs}
    case "self":
        return "created_by = ?", []interface{}{userID}
    case "custom":
        // 查询自定义部门列表
        deptIDs := getCustomDepartmentIDs(userID, resource)
        return "department_id IN (?)", []interface{}{deptIDs}
    default:
        return "1 = 0", nil  // 默认无权限
    }
}
```

### 5.3 前端权限控制

**菜单级**：根据用户权限动态生成侧边栏菜单。

**按钮级**：通过自定义指令或权限组件控制按钮显示/隐藏。

```tsx
// 权限按钮组件
interface PermissionButtonProps {
  permission: string;  // 权限编码，如 'user:create'
  children: React.ReactNode;
}

const PermissionButton: React.FC<PermissionButtonProps> = ({ permission, children }) => {
  const hasPermission = usePermission(permission);
  if (!hasPermission) return null;
  return <>{children}</>;
};

// 使用
<PermissionButton permission="member:create">
  <Button type="primary">新增会员</Button>
</PermissionButton>
```

---

## 6. 系统配置与权限开关

### 6.1 系统配置表 (configs)

所有可配置的功能开关、参数都存在数据库中，支持动态修改。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `config_key` | VARCHAR(100) | 配置键（唯一） |
| `config_value` | TEXT | 配置值（JSON 格式，支持复杂类型） |
| `config_type` | VARCHAR(20) | 类型：string/number/boolean/json |
| `description` | VARCHAR(255) | 配置说明 |
| `category` | VARCHAR(50) | 配置分类 |
| `is_public` | BOOLEAN | 是否公开（前端可读取） |
| `updated_by` | UUID | 更新人 |
| `updated_at` | TIMESTAMP | 更新时间 |
| `created_at` | TIMESTAMP | 创建时间 |

### 6.2 权限开关配置示例

| config_key | 类型 | 默认值 | 说明 |
|-----------|------|--------|------|
| `feature.internship.enabled` | boolean | true | 是否启用实习管理模块 |
| `feature.internship.allow_self_edit` | boolean | true | 是否允许学生自行修改实习信息 |
| `feature.internship.dept_leader_can_edit` | boolean | true | 部门负责人是否可以修改部门内实习信息 |
| `feature.meeting.vote.enabled` | boolean | true | 是否启用会议投票 |
| `feature.meeting.weighted_vote.enabled` | boolean | false | 是否启用加权投票（二期） |
| `feature.notification.email.enabled` | boolean | true | 是否启用邮件通知 |
| `feature.notification.wechat.enabled` | boolean | false | 是否启用微信通知（二期） |
| `approval.member.requires_interview` | boolean | false | 会员申请是否需要面试 |
| `approval.interview.rounds` | number | 2 | 干事面试轮数（1或2） |

### 6.3 配置中心服务

```go
// ConfigService 配置服务接口
type ConfigService interface {
    Get(key string) (string, error)
    GetBool(key string) (bool, error)
    GetInt(key string) (int, error)
    GetJSON(key string, out interface{}) error
    Set(key string, value interface{}, updatedBy uuid.UUID) error
    ListByCategory(category string) ([]Config, error)
    GetAllPublic() (map[string]interface{}, error)
}
```

前端通过 `/api/v1/configs/public` 接口获取所有公开配置，用于控制页面功能的显示/隐藏。

---

## 7. 预置角色与权限

### 7.1 系统内置角色

| 角色编码 | 角色名称 | 说明 |
|---------|---------|------|
| `super_admin` | 超级管理员 | 拥有所有权限，系统初始化时创建 |
| `president` | 会长 | 协会最高权限，可管理所有模块 |
| `vice_president` | 副会长 | 协助会长，权限仅次于会长 |
| `minister` | 部长 | 可管理本部门成员和事务 |
| `vice_minister` | 副部长 | 协助部长 |
| `secretary` | 干事 | 普通干事权限 |
| `member` | 会员 | 普通会员，仅可查看个人信息和申请 |

### 7.2 权限编码规范

```
{资源}:{操作}

示例：
user:create        创建用户
user:read          查看用户
user:update        更新用户
user:delete        删除用户
member:approve     审批入会申请
interview:score    面试评分
meeting:vote       会议投票
task:assign        任务分配
workflow:design    流程设计
stats:view         查看统计报表
```

---

## 8. API 接口

### 8.1 角色管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/roles` | 分页查询角色列表 |
| GET | `/api/v1/roles/:id` | 获取角色详情 |
| POST | `/api/v1/roles` | 创建角色 |
| PUT | `/api/v1/roles/:id` | 更新角色 |
| DELETE | `/api/v1/roles/:id` | 删除角色 |
| GET | `/api/v1/roles/:id/permissions` | 获取角色权限列表 |
| PUT | `/api/v1/roles/:id/permissions` | 分配角色权限 |
| GET | `/api/v1/roles/:id/users` | 获取角色下的用户列表 |

### 8.2 权限管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/permissions/tree` | 获取权限树 |
| GET | `/api/v1/permissions/:id` | 获取权限详情 |
| POST | `/api/v1/permissions` | 创建权限 |
| PUT | `/api/v1/permissions/:id` | 更新权限 |
| DELETE | `/api/v1/permissions/:id` | 删除权限 |

### 8.3 用户-角色管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/users/:id/roles` | 获取用户角色列表 |
| PUT | `/api/v1/users/:id/roles` | 分配用户角色 |

### 8.4 系统配置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/configs` | 分页查询配置列表 |
| GET | `/api/v1/configs/public` | 获取公开配置（前端用） |
| PUT | `/api/v1/configs/:key` | 更新配置 |
| PUT | `/api/v1/configs/batch` | 批量更新配置 |

---

## 9. 前端权限管理页面

### 9.1 角色管理页面
- 角色列表（树形展示，支持角色继承）
- 新增/编辑角色
- 分配权限（树形权限勾选 + 数据权限范围选择）
- 查看角色下的用户

### 9.2 权限管理页面
- 权限树（菜单/按钮/接口分类展示）
- 新增/编辑权限
- 权限排序

### 9.3 系统配置页面
- 按分类展示配置项
- 编辑配置值（支持文本、数字、布尔、JSON 等类型）
- 配置分组：功能开关、通知设置、审批设置、其他

---

## 10. 注意事项

1. **超级管理员权限绕过**：`super_admin` 角色拥有所有权限，权限校验时直接放行，不查权限表
2. **权限缓存**：用户权限信息缓存在 Redis 中，减少数据库查询。权限变更时主动失效缓存
3. **数据权限性能**：数据权限的部门子树查询使用闭包表或路径枚举优化
4. **前端权限安全**：前端的按钮隐藏只是体验优化，后端必须做严格的权限校验，不能依赖前端
5. **配置缓存**：系统配置缓存在内存 + Redis 中，更新时主动刷新
