# Step 输出契约

每个 step 执行完成后，无论是 subagent 模式还是内联模式，都必须产出以下标准化结果。

## 输出结构

每步完成时，通过更新 `hap-context.json` 来表达结果：

### 成功

1. 将 `progress` 字段更新为当前步骤对应的完成状态（见 `PROGRESS.md`）
2. 写入该步骤负责的 ID 映射字段（见 `CONTEXT.md`）
3. 通过 ⛔ 验证断言

### 失败

1. **不更新** `progress` 字段（保持前一步的状态值）
2. 调度器发现 progress 未推进 → 判定为失败 → 向用户报告错误

## 调度器验证方法

调度器在每步完成后，读取 `hap-context.json` 检查：

```
if context.progress == 当前步骤的预期完成状态:
    ✅ 步骤成功，继续下一步
else:
    ❌ 步骤失败，报告错误并停止
```

## 各步骤的预期产出

| Step | 预期 progress | 必须写入的字段 |
|------|--------------|---------------|
| 1 | `app_created` | `appId`, `sectionIdByName` |
| 2 | `worksheets_created` | `worksheetIdByName` |
| 3 | `fields_refreshed` | _(写入 worksheetContext.json)_ |
| 4 | `actions_created` | `actionIdByName` |
| 5 | `views_created` | `viewIdByName` |
| 6 | `sample_data_created` | _(无新字段，数据直接写入平台)_ |
| 7 | `pages_created` | `customPageIdByName`, `chatbotIdByName`（若有） |
| 8 | `roles_created` | `roleContext` |
| 9 | `workflows_designed` | _(工作流设计写入 hap-plan.json)_ |
| 10 | `system_workflows_published` | _(系统工作流 processId 已 publish)_ |
| 11 | `completed` | `customActionWorkflows` |
