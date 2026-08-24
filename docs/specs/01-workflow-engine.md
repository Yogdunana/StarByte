# 流程引擎设计文档

> **版本**: v1.0
> **日期**: 2026-08-24
> **模块**: workflow
> **状态**: 一期设计

---

## 1. 设计目标

流程引擎是 StarByte 系统的核心底层能力，为面试、入会申请、任务审批等所有需要审批流转的业务场景提供支撑。

设计原则：
- **插件化**：节点类型可扩展，新增节点无需修改核心代码
- **事件驱动**：流程状态变更通过事件通知外部系统，解耦业务逻辑
- **版本化**：流程定义支持多版本，运行中的实例不受定义变更影响
- **可视化**：支持拖拽式流程设计，实时预览流程运行状态
- **可扩展**：预留充分的扩展点，二期功能可平滑接入

---

## 2. 核心概念

### 2.1 流程定义 (Flow Definition)
流程的"模板"，定义了流程有哪些节点、节点之间如何连接、每个节点的配置。支持多版本管理。

### 2.2 流程实例 (Flow Instance)
流程的一次"运行"。当某个业务（如面试申请）启动流程时，会基于某个版本的流程定义创建一个流程实例。

### 2.3 流程任务 (Flow Task)
流程实例中的待办事项。当流程流转到审批节点时，会为审批人生成对应的任务。

### 2.4 流程节点 (Flow Node)
流程中的一个步骤。每个节点有类型、配置、输入输出。

### 2.5 流程变量 (Flow Variable)
流程实例运行过程中的数据，如申请人信息、审批意见、表单数据等。条件分支根据变量值判断走向。

---

## 3. 节点类型体系

### 3.1 节点分类

| 分类 | 节点类型 | 说明 | 一期实现 |
|------|---------|------|---------|
| **流程控制** | `start` | 开始节点 | ✅ |
| | `end` | 结束节点 | ✅ |
| | `exclusive_gateway` | 排他网关（条件分支，多选一） | ✅ |
| | `parallel_gateway` | 并行网关（所有分支同时执行） | ✅ |
| **人工任务** | `approval` | 审批节点 | ✅ |
| | `sign_task` | 签字/确认节点 | 🔲 二期 |
| **自动任务** | `service_task` | 服务任务（调用 API / 执行业务逻辑） | ✅ |
| | `notification_task` | 通知任务（发送消息） | ✅ |
| **高级** | `sub_process` | 子流程 | 🔲 二期 |
| | `timer_event` | 定时事件（超时自动处理） | 🔲 二期 |

### 3.2 审批节点配置

```json
{
  "nodeType": "approval",
  "config": {
    "approvalType": "single",          // single(单人) / all(会签) / any(或签) / ratio(比例)
    "assigneeStrategy": "static",      // static(指定人) / role(指定角色) / dept_leader(部门负责人) / initiator(发起人)
    "assignees": [],                    // 指定审批人ID列表
    "roleId": "",                       // 指定角色ID
    "passRatio": 0,                     // 通过比例（approvalType=ratio 时有效）
    "allowReject": true,                // 是否允许驳回
    "allowTransfer": true,              // 是否允许转办
    "allowAddSign": true,               // 是否允许加签
    "allowRollback": true,              // 是否允许退回
    "formFields": []                    // 审批表单字段配置
  }
}
```

### 3.3 条件分支节点配置

```json
{
  "nodeType": "exclusive_gateway",
  "config": {
    "branches": [
      {
        "id": "branch_1",
        "label": "面试通过",
        "expression": "interview.score >= 60"
      },
      {
        "id": "branch_2",
        "label": "面试不通过",
        "expression": "interview.score < 60",
        "isDefault": true
      }
    ]
  }
}
```

**表达式引擎**：基于 [govaluate](https://github.com/Knetic/govaluate)，支持：
- 算术运算：`+ - * / %`
- 比较运算：`== != > >= < <=`
- 逻辑运算：`&& || !`
- 字符串操作：`+` 拼接、`contains()` 等
- 变量访问：通过 `.` 访问嵌套属性，如 `applicant.name`

---

## 4. 数据模型

### 4.1 flow_definitions（流程定义表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `key` | VARCHAR(100) | 流程标识（业务用，如 `interview_flow`） |
| `name` | VARCHAR(200) | 流程名称 |
| `description` | TEXT | 流程描述 |
| `category` | VARCHAR(50) | 分类（interview/member/task/custom） |
| `status` | SMALLINT | 状态：0=草稿 1=已发布 2=已停用 |
| `created_by` | UUID | 创建人 |
| `updated_by` | UUID | 更新人 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

### 4.2 flow_definition_versions（流程定义版本表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `definition_id` | UUID | 流程定义ID |
| `version` | INT | 版本号（从1开始递增） |
| `bpmn_data` | JSONB | 流程定义数据（节点、连线等完整结构） |
| `status` | SMALLINT | 状态：0=历史 1=当前版本 |
| `published_by` | UUID | 发布人 |
| `published_at` | TIMESTAMP | 发布时间 |
| `created_at` | TIMESTAMP | 创建时间 |

### 4.3 flow_instances（流程实例表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `definition_id` | UUID | 流程定义ID |
| `definition_version_id` | UUID | 流程定义版本ID（启动时的版本） |
| `business_key` | VARCHAR(100) | 业务键（关联业务数据，如面试申请ID） |
| `business_type` | VARCHAR(50) | 业务类型 |
| `initiator_id` | UUID | 发起人ID |
| `status` | SMALLINT | 状态：0=运行中 1=已完成 2=已终止 3=已挂起 |
| `current_node_ids` | JSONB | 当前节点ID列表（并行时多个） |
| `started_at` | TIMESTAMP | 启动时间 |
| `ended_at` | TIMESTAMP | 结束时间 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

### 4.4 flow_tasks（流程任务表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `instance_id` | UUID | 流程实例ID |
| `node_id` | VARCHAR(100) | 节点ID（对应流程定义中的节点） |
| `node_name` | VARCHAR(200) | 节点名称 |
| `task_type` | VARCHAR(50) | 任务类型：approval/notification/service |
| `assignee_id` | UUID | 处理人ID |
| `status` | SMALLINT | 状态：0=待处理 1=已通过 2=已驳回 3=已转办 4=已撤回 5=已取消 |
| `action` | VARCHAR(50) | 操作：approve/reject/transfer/rollback |
| `comment` | TEXT | 审批意见 |
| `form_data` | JSONB | 表单数据 |
| `due_date` | TIMESTAMP | 截止时间 |
| `claimed_at` | TIMESTAMP | 签收时间 |
| `completed_at` | TIMESTAMP | 完成时间 |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

### 4.5 flow_histories（流程历史记录表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `instance_id` | UUID | 流程实例ID |
| `task_id` | UUID | 任务ID（可为空，如开始/结束节点） |
| `node_id` | VARCHAR(100) | 节点ID |
| `node_name` | VARCHAR(200) | 节点名称 |
| `node_type` | VARCHAR(50) | 节点类型 |
| `operator_id` | UUID | 操作人ID |
| `action` | VARCHAR(50) | 操作类型：start/approve/reject/transfer/rollback/complete |
| `comment` | TEXT | 操作意见 |
| `from_node_id` | VARCHAR(100) | 来源节点ID |
| `to_node_id` | VARCHAR(100) | 目标节点ID |
| `created_at` | TIMESTAMP | 操作时间 |

### 4.6 flow_variables（流程变量表）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 主键 |
| `instance_id` | UUID | 流程实例ID |
| `key` | VARCHAR(100) | 变量名 |
| `value` | JSONB | 变量值 |
| `scope` | VARCHAR(20) | 作用域：global/local |
| `created_at` | TIMESTAMP | 创建时间 |
| `updated_at` | TIMESTAMP | 更新时间 |

---

## 5. 引擎核心架构

### 5.1 核心接口定义

```go
// NodeHandler 节点处理器接口，所有节点类型都必须实现
type NodeHandler interface {
    // Type 返回节点类型标识
    Type() string
    
    // Execute 执行节点逻辑，返回输出的连线ID列表
    Execute(ctx context.Context, instance *FlowInstance, node *FlowNode, vars map[string]interface{}) ([]string, error)
    
    // OnEnter 节点进入时的回调（如创建审批任务）
    OnEnter(ctx context.Context, instance *FlowInstance, node *FlowNode, vars map[string]interface{}) error
    
    // OnLeave 节点离开时的回调
    OnLeave(ctx context.Context, instance *FlowInstance, node *FlowNode, vars map[string]interface{}) error
    
    // Validate 验证节点配置是否合法
    Validate(node *FlowNode) error
}

// TaskHandler 任务处理器接口
type TaskHandler interface {
    // Complete 完成任务
    Complete(ctx context.Context, task *FlowTask, action TaskAction, comment string, formData map[string]interface{}) error
    
    // Transfer 转办任务
    Transfer(ctx context.Context, task *FlowTask, targetUserID uuid.UUID, comment string) error
    
    // Rollback 退回任务
    Rollback(ctx context.Context, task *FlowTask, targetNodeID string, comment string) error
    
    // Withdraw 撤回任务（发起人撤回）
    Withdraw(ctx context.Context, task *FlowTask, comment string) error
}
```

### 5.2 节点注册表

```go
// NodeRegistry 节点处理器注册表
type NodeRegistry struct {
    handlers map[string]NodeHandler
}

// Register 注册节点处理器
func (r *NodeRegistry) Register(handler NodeHandler) {
    r.handlers[handler.Type()] = handler
}

// Get 获取节点处理器
func (r *NodeRegistry) Get(nodeType string) (NodeHandler, bool) {
    h, ok := r.handlers[nodeType]
    return h, ok
}
```

### 5.3 流程引擎核心接口

```go
// FlowEngine 流程引擎接口
type FlowEngine interface {
    // Start 启动流程实例
    Start(ctx context.Context, definitionKey string, businessKey string, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*FlowInstance, error)
    
    // CompleteTask 完成任务
    CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action TaskAction, comment string, formData map[string]interface{}) error
    
    // TransferTask 转办任务
    TransferTask(ctx context.Context, taskID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, comment string) error
    
    // RollbackTask 退回任务
    RollbackTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, targetNodeID string, comment string) error
    
    // Terminate 终止流程
    Terminate(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error
    
    // Suspend 挂起流程
    Suspend(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error
    
    // Resume 恢复流程
    Resume(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error
}
```

---

## 6. 流程流转机制

### 6.1 启动流程

```
调用 Start()
    ↓
根据 definitionKey 查找当前版本的流程定义
    ↓
创建 flow_instances 记录（status=0 运行中）
    ↓
初始化流程变量（variables 参数）
    ↓
找到 start 节点
    ↓
触发 FlowStartedEvent 事件
    ↓
进入节点：执行 node.OnEnter()
    ↓
执行节点：执行 node.Execute()，获取输出连线
    ↓
离开节点：执行 node.OnLeave()
    ↓
沿输出连线到达下一个节点
    ↓
重复进入/执行/离开，直到遇到等待节点（如审批）或结束节点
```

### 6.2 完成审批任务

```
调用 CompleteTask()
    ↓
校验任务状态和操作权限
    ↓
更新 flow_tasks 状态（status=已通过/已驳回）
    ↓
写入 flow_histories 记录
    ↓
更新流程变量（form_data 合并到 variables）
    ↓
触发 TaskCompletedEvent 事件
    ↓
执行节点的 Execute()，获取输出连线
    ↓
沿连线继续流转到下一个节点
```

### 6.3 并行网关

- **进入并行网关（fork）**：激活所有输出分支，每个分支独立流转
- **汇聚并行网关（join）**：等待所有输入分支都到达后，才继续向下流转

---

## 7. 事件系统

### 7.1 事件类型

| 事件 | 触发时机 | 携带数据 |
|------|---------|---------|
| `FlowStartedEvent` | 流程启动时 | 实例ID、定义ID、发起人 |
| `FlowCompletedEvent` | 流程正常结束时 | 实例ID、结束时间 |
| `FlowTerminatedEvent` | 流程被终止时 | 实例ID、终止原因、操作人 |
| `TaskCreatedEvent` | 任务创建时 | 任务ID、处理人、实例ID |
| `TaskCompletedEvent` | 任务完成时 | 任务ID、操作、意见 |
| `TaskAssignedEvent` | 任务分配/转办时 | 任务ID、原处理人、新处理人 |
| `NodeEnteredEvent` | 进入节点时 | 实例ID、节点ID |
| `NodeLeftEvent` | 离开节点时 | 实例ID、节点ID |

### 7.2 事件订阅者

- **通知服务**：监听任务创建事件，发送站内消息 + WebSocket + 邮件
- **审计服务**：监听所有事件，记录审计日志
- **统计服务**：监听流程/任务事件，更新统计数据
- **业务回调**：业务模块可订阅事件，在流程特定节点执行业务逻辑

---

## 8. 前端流程设计器

### 8.1 技术选型

- **核心库**：React Flow 11.x
- **拖拽面板**：自定义节点类型，从左侧面板拖拽到画布
- **属性面板**：右侧配置面板，根据选中节点类型动态显示配置项
- **工具栏**：缩放、保存、发布、预览、撤销/重做

### 8.2 数据格式

流程定义数据（存储在 `bpmn_data` 字段中）：

```json
{
  "nodes": [
    {
      "id": "node_start",
      "type": "start",
      "position": { "x": 100, "y": 200 },
      "data": { "label": "开始" }
    },
    {
      "id": "node_approval_1",
      "type": "approval",
      "position": { "x": 300, "y": 200 },
      "data": {
        "label": "一面",
        "config": {
          "approvalType": "single",
          "assigneeStrategy": "role",
          "roleId": "interviewer-role-id"
        }
      }
    }
  ],
  "edges": [
    {
      "id": "edge_1",
      "source": "node_start",
      "target": "node_approval_1",
      "animated": true
    }
  ],
  "viewport": {
    "x": 0,
    "y": 0,
    "zoom": 1
  }
}
```

### 8.3 设计器功能列表

| 功能 | 说明 | 一期 |
|------|------|------|
| 节点拖拽 | 从左侧面板拖入节点 | ✅ |
| 连线 | 从节点拖出连线到另一节点 | ✅ |
| 删除节点/连线 | 删除选中的元素 | ✅ |
| 属性配置 | 右侧面板配置节点属性 | ✅ |
| 保存草稿 | 保存为草稿状态 | ✅ |
| 发布 | 发布新版本 | ✅ |
| 版本历史 | 查看历史版本 | ✅ |
| 流程预览 | 模拟流程运行 | 🔲 二期 |
| 撤销/重做 | Ctrl+Z / Ctrl+Y | ✅ |
| 放大缩小 | 画布缩放 | ✅ |
| 适配画布 | 一键适配到画布大小 | ✅ |
| 网格对齐 | 节点吸附网格 | ✅ |
| 复制/粘贴 | 复制节点 | ✅ |

---

## 9. API 接口

### 9.1 流程定义管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/workflow/definitions` | 分页查询流程定义列表 |
| GET | `/api/v1/workflow/definitions/:id` | 获取流程定义详情 |
| POST | `/api/v1/workflow/definitions` | 创建流程定义（草稿） |
| PUT | `/api/v1/workflow/definitions/:id` | 更新流程定义 |
| DELETE | `/api/v1/workflow/definitions/:id` | 删除流程定义 |
| POST | `/api/v1/workflow/definitions/:id/publish` | 发布流程定义 |
| GET | `/api/v1/workflow/definitions/:id/versions` | 获取版本列表 |
| GET | `/api/v1/workflow/definitions/:id/versions/:versionId` | 获取指定版本详情 |

### 9.2 流程实例管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/workflow/instances` | 启动流程实例 |
| GET | `/api/v1/workflow/instances` | 分页查询流程实例 |
| GET | `/api/v1/workflow/instances/:id` | 获取流程实例详情 |
| GET | `/api/v1/workflow/instances/:id/diagram` | 获取流程实例状态图数据 |
| POST | `/api/v1/workflow/instances/:id/terminate` | 终止流程 |
| POST | `/api/v1/workflow/instances/:id/suspend` | 挂起流程 |
| POST | `/api/v1/workflow/instances/:id/resume` | 恢复流程 |

### 9.3 流程任务管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/workflow/tasks/todo` | 获取我的待办任务 |
| GET | `/api/v1/workflow/tasks/done` | 获取我的已办任务 |
| GET | `/api/v1/workflow/tasks/:id` | 获取任务详情 |
| POST | `/api/v1/workflow/tasks/:id/approve` | 审批通过 |
| POST | `/api/v1/workflow/tasks/:id/reject` | 审批驳回 |
| POST | `/api/v1/workflow/tasks/:id/transfer` | 转办任务 |
| POST | `/api/v1/workflow/tasks/:id/rollback` | 退回任务 |
| POST | `/api/v1/workflow/tasks/:id/withdraw` | 撤回任务（发起人） |

### 9.4 流程历史

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/workflow/instances/:id/history` | 获取流程历史记录 |

---

## 10. 一期实施范围

### 10.1 后端
- ✅ 流程引擎核心（启动、流转、结束）
- ✅ 节点：开始、结束、审批、排他网关、并行网关、通知任务、服务任务
- ✅ 审批模式：单人、会签、或签、比例通过
- ✅ 任务操作：通过、驳回、转办、退回、撤回
- ✅ 表达式引擎（条件分支）
- ✅ 事件系统 + 事件总线
- ✅ 流程定义 CRUD + 版本管理
- ✅ 流程实例管理
- ✅ 任务管理（待办/已办）
- ✅ 流程历史记录
- ✅ 与面试模块集成

### 10.2 前端
- ✅ 流程列表页
- ✅ 流程设计器（React Flow）
- ✅ 节点属性配置面板
- ✅ 流程发布/版本管理
- ✅ 我的待办/已办列表
- ✅ 审批操作弹窗
- ✅ 流程监控图（当前节点高亮）
- ✅ 流程历史记录查看

### 10.3 二期规划
- 🔲 子流程
- 🔲 定时事件 / 超时处理
- 🔲 流程模拟预览
- 🔲 会签加签
- 🔲 更多节点类型（签字、触发器等）
- 🔲 入会申请、任务流转接入流程引擎
- 🔲 流程监控大屏
