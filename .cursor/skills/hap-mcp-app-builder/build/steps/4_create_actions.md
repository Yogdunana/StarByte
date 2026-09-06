# Step 4：创建自定义动作

你是 HAP 自定义动作搭建专家。根据方案中每张工作表的 `customActions` 列表，批量创建自定义动作按钮。

## 输入数据

- `appId`：应用 ID
- `worksheetContext`：工作表结构列表（含字段 alias/ID 映射），来自 `worksheetContext.json`（只读）
- `worksheetCustomActions`：来自 `hap-plan.json` 的自定义动作规划

## 执行流程

对每张有自定义动作的工作表：

1. 调用 `create_custom_actions` 批量创建该表的所有动作
2. 记录返回的 `actionIdByName`（格式：`"工作表名/动作名" → actionId`）
3. **关键**：`type=triggerWorkflow` 的动作，系统会自动创建工作流外壳并返回 `processId`——必须记录到 `customActionWorkflows[]`
4. 更新 `hap-context.json`：写入 `actionIdByName`、`customActionWorkflows`（不写 `progress`，由调度器统一管理）

**⛔ 验证断言**：`actionIdByName` 条目数 = plan 中全部自定义动作总数。`customActionWorkflows[]` 条目数 = plan 中 `type=triggerWorkflow` 的动作数。

---

## build阶段需要补充的字段

| type | 说明 | 必填额外参数 |
|---|---|---|
| `updateCurrentRecord` | 允许用户填写当前记录的指定字段 | `updateFields`（字段 alias 或 ID 列表）|
| `createRelatedRecord` | 在关联表中新建一条关联记录 | `relationField`（关联字段 alias 或 ID）|
| `triggerWorkflow` | 直接触发绑定的工作流 | 无 |

> 所有类型均可选配 `enableWhen`（触发条件），详见下方 enableWhen 章节。

### updateFields 推断

plan 的 `description` 描述了"用户填什么"，build 阶段从 worksheetFields 里挑出对应字段的 alias 列表。

例：
- plan: `"办理签到"` description "填写实际到场时间、证件核验情况、访客证编号"
- worksheetFields 含: `实际签到时间(alias=biz_check_in_time)` / `证件核验通过(alias=biz_id_verified)` / `访客证编号(alias=biz_badge_no)`
- → `updateFields: ["biz_check_in_time", "biz_id_verified", "biz_badge_no"]`

### relationField 推断

plan 的 `targetWorksheet` 告诉我们要建哪张表的关联记录。worksheetFields 里找到 type=Relation 且 dataSource 指向 targetWorksheet 的字段，取其 alias。

例：
- plan: `"登记异常"` type=createRelatedRecord, targetWorksheet="来访异常"
- worksheetFields 含: `相关异常(type=Relation, alias=biz_exception_relation, dataSource=来访异常的 ID)`
- → `relationField: "biz_exception_relation"`

## enableWhen（触发条件）

满足条件时按钮才可用，结构与视图 `filter` 完全相同（最外层必须是 group）。不传则始终可用。

> ⚠️ **必须根据业务逻辑判断是否设置 `enableWhen`，不要默认省略。**
>
> - 若按钮有前置状态要求（如"借书"要求图书状态为"在库"、"发货"要求订单状态为"已付款"），**必须设置 `enableWhen`**
> - 只有真正无前置条件的操作（如"添加备注"、"发送通知"）才可不设

**示例 — 借书按钮（仅当图书状态为"在库"时可用）**：
```json
"enableWhen": {
  "type": "group",
  "logic": "AND",
  "children": [
    { "type": "condition", "field": "status", "operator": "eq", "value": "在库" }
  ]
}
```

**示例 — 确认归还按钮（当借阅状态为"在借中"或"已超期"时可用）**：
```json
"enableWhen": {
  "type": "group",
  "logic": "AND",
  "children": [
    { "type": "condition", "field": "borrowStatus", "operator": "in", "value": ["在借中", "已超期"] }
  ]
}
```

> [!CAUTION]
> **同一字段匹配多个值 = 单个 `in` condition + value 数组。严禁拆成多个 condition 用 AND 组合（逻辑上永远不成立）。**

---

## 字段引用规则

- 优先使用字段 `alias`，无 alias 时用字段 ID
- 只能使用 `worksheetContext` 中提供的字段，不猜测字段
- `updateFields` 中的字段必须是用户实际需要填写的字段
- `relationField` 必须是已存在的关联字段

---

## 创建结果

`create_custom_actions` 返回 `actionIdByName`（按钮名称 → actionId 的映射），供后续视图的 `actions` 配置引用。

对 `type=triggerWorkflow` 的动作，还需记录：
```json
{
  "name": "动作名称",
  "processId": "系统返回的 processId",
  "worksheetName": "所在工作表名称",
  "intentHints": "来自 plan 的业务意图"
}
```
这些工作流在后续工作流阶段需要填充节点，不需要重新调用 `create_process`。
