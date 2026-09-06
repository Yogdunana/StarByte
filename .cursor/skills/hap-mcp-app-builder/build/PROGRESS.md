# Progress 状态机

`hap-context.json` 中的 `progress` 字段是一个有限状态机。**所有 progress 写入均由调度器完成**，各 step 只负责写入自己的产出数据（ID 映射等），不写 progress。

## 状态定义

| 状态值 | 含义 | 前置状态 | 写入时机 |
|--------|------|---------|----------|
| _(空/不存在)_ | 初始状态 | — | — |
| `app_created` | 应用和导航分组已创建 | _(空)_ | Step 1 完成后 |
| `worksheets_created` | 所有工作表已创建 | `app_created` | Step 2 完成后 |
| `fields_refreshed` | 字段结构已刷新 | `worksheets_created` | Step 3 完成后 |
| `actions_created` | 自定义动作已创建 | `fields_refreshed` | Step 4 完成后 |
| `views_created` | 视图已创建 | `actions_created` | Step 5 完成后 |
| `sample_data_created` | 示例数据已写入 | `views_created` | Step 5 和 Step 6 都完成后 |
| `page_shells_created` | 页面空壳与 AI 助手已创建 | `sample_data_created` | Step 5b 完成后 |
| `config_completed` | 页面组件 + 角色 + 工作流设计全部完成 | `page_shells_created` | Step 7、Step 8、Step 9 全部完成后 |
| `workflows_deployed` | 系统工作流 + 自定义动作工作流均已发布 | `config_completed` | Step 10 和 Step 11 都完成后 |
| `completed` | 全部完成（含 CLI 建后精修对账） | `workflows_deployed` | Step 12 完成后 |
| `failed` | 执行失败 | _(任意)_ | 任意步骤失败时 |

> **并行汇合点**：`page_shells_created` → `config_completed` 之间，Step 7/8/9 三路并行。调度器内部追踪各路完成状态，三路全部完成后才写入 `config_completed`。

## 规则

- **只能前进不能后退**——每步只能将 progress 推进到下一个状态
- **不能跳跃**——不允许从 `app_created` 直接跳到 `views_created`
- **failed 是终态**——进入 `failed` 后必须人工介入
- **调度器独占写入**——各 step 不写 progress，由调度器在验证通过后统一写入
- **并行汇合由调度器追踪**——并行步骤各自完成后调度器在内存中记录，全部完成才推进 progress
