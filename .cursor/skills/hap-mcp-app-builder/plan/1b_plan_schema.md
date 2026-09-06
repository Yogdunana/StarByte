# HapPlan JSON 结构规范

用户已确认方案总览后，按以下规则生成结构化 JSON，写入 `hap-plan.json`。

> **⚠️ 格式红线**：所有字符串参数中，**禁止使用未转义的英文双引号 `"`**。若需引用文本，必须使用中文双引号 `""` 或单引号 `''`。

---

## 顶层结构

```json
{
  "org_id": "来自 Step 0 选择的组织 ID",
  "appName": "应用名称",
  "appIcon": "图标名称，如 0_lego",
  "appColor": "主题色，如 #2196F3",
  "navLayout": "group | tree | top | card",
  "navColor": "appColor | white | gray | black",
  "enableExternalPortal": true,
  "worksheets": [],
  "worksheetViews": [],
  "worksheetCustomActions": [],
  "customPages": [],
  "workflows": [],
  "roles": [],
  "aiAssistants": [],
  "navGroups": []
}
```

### 属性约束说明：
* **`org_id`**：组织 ID，键名必须使用下划线风格 `org_id`，**严禁写成驼峰 `orgId`**（MCP 工具参数名为 `org_id`）。
* **`appIcon`**：应用图标，必须严格从 `plan/icon_and_style_guide.md` 中挑选契合业务语义的预设图标。
* **`appColor`**：应用主题色，必须且只能使用 `plan/icon_and_style_guide.md` 中提供的 9 种官方高端 Hex 色值。**只有应用本身有颜色**，工作表、自定义页面、AI 助手均无颜色属性。
* **`navColor`**：导航栏颜色，若使用 `appColor` 代表导航栏使用应用主题色。
* **`enableExternalPortal`**：是否启用外部门户。`true` 表示启用外部门户（若规划了外部门户角色 `"roleScope": "externalPortal"`，此项必须为 `true`）；`false` 或不传表示不启用。

---

## 一、`worksheets`

每张工作表必须包含：
- `icon`：从 `plan/icon_and_style_guide.md` 中挑选最契合业务语义的图标。不同工作表应尽量使用不同图标，避免全部雷同
- `description`：一句话说明该表的业务定位与核心用途（如「记录每一笔图书借阅与归还事件」）

### `fields`（紧凑字符串数组）

### 基本格式

1. 每个字段用紧凑格式：`"字段名(Type)"`，如 `"任务标题(Text)"`、`"负责人(Collaborator)"`
2. 关联字段：`"字段名(Relation:目标工作表名)"`，如 `"关联图书(Relation:图书)"`
3. 自关联：`"字段名(selfRelation)"`，如 `"上级分类(selfRelation)"`
4. **SingleSelect / MultipleSelect / Dropdown 必须携带选项值**，用 `/` 分隔：`"状态(SingleSelect:待处理/处理中/已完成/已逾期)"`。禁止只写 `"状态(SingleSelect)"` 而不列选项
5. 必须充分利用专属字段类型：坐标/定位 → `Location`，行政区划 → `Region`，流水号 → `AutoNumber`，人员 → `Collaborator`，部门 → `Department`，手机号 → `PhoneNumber`，邮箱 → `Email`，金额/价格 → `Currency`，日期计算 → `DateFormula`（如"应还日期(DateFormula)"）。**严禁将这些字段降级为 `Text` 或 `Number`**

### 字段丰富度要求（核心）

> ⚠️ **字段不够丰富是最常见的设计缺陷。** Plan 阶段输出的字段就是最终要创建的字段清单，Build 阶段不应大量补充。因此必须在此阶段就输出完整、丰富的字段。

**最低字段数量**：
- 主数据表（如客户、图书、商品、设备）：**≥ 15 个字段**
- 业务单据表（如订单、借阅记录、工单）：**≥ 12 个字段**
- 过程/流水表（如入库单、巡检记录）：**≥ 8 个字段**

**字段维度覆盖**：按 `plan/design_guide.md` 中「字段完整性」的 7 个维度和「业务场景字段补全」逐项检查，确保每个维度至少覆盖 1 个字段，并根据业务场景主动补齐容易遗漏的字段。

### `consistencyNotes`（必填）

每张表必须填写字段与视图/工作流/自定义动作的一致性说明：
- 哪些 SingleSelect 字段的哪些选项值支撑哪些视图的筛选条件
- 哪些字段会被工作流读取或更新
- 自定义动作的前后状态变化
- 示例：`"状态字段的已逾期选项用于超期事项视图筛选；点击处理按钮时状态须为待处理，工作流执行后更新为处理中；超期检查工作流自动将状态更新为已逾期"`

## 二、`worksheetViews`（独立顶层数组）

1. 每项包含 `worksheet`（工作表名称）和 `views`（紧凑字符串数组）
2. 视图格式：`"视图名(Type)"`，如 `"列表视图(Table)"`、`"看板视图(Kanban)"`
3. 为每张表优先生成 1～4 个最有业务价值的核心视图
4. **视图-字段-工作流三方闭环**：所有视图筛选条件和时效性状态必须通过闭环检查后才能提交

## 三、`worksheetCustomActions`（独立顶层数组）

1. 每项包含 `worksheet`（工作表名称）和 `actions`（动作对象数组）
2. 动作对象字段：
   - `name`：动作名称
   - `description`：由谁在什么场景下点击，以及需要填写什么
   - `type`：`"updateCurrentRecord"` / `"createRelatedRecord"` / `"triggerWorkflow"`
   - `targetWorksheet`：仅 `type="createRelatedRecord"` 时填目标工作表名
   - `relateFieldName`：仅 `type="createRelatedRecord"` 时填，表示源工作表中用于物理关联目标表的关联字段名称（例如当前工作表中存在 `"关联商机(Relation:销售机会)"` 字段，此处则必须填写 `"关联商机"`）
   - `enableCondition`（可选）：按钮的前置状态条件，自然语言描述。有前置状态要求的动作必填
   - `intentHints`：`type="triggerWorkflow"` 时必须填写业务效果与约束数组（`[{label}]`）

3. 只生成业务上真正需要**人工触发**的动作
4. **类型自检**：若 `updateCurrentRecord` 的所有字段值均可由系统自动确定，则改为 `triggerWorkflow`
5. **审批自检（严禁违反）**：严禁创建「审批通过」「审批驳回」「审批否决」等直接操作审批结果的自定义动作。正确做法：创建一个 `triggerWorkflow` 类型的动作（如「提交审批」「发起审批」），或通过新建记录时自动触发（`worksheet_event`），由工作流内部的审批业务块 (`approval_block`) 来处理通过/驳回/填写等流转——通过与驳回由审批人在审批节点中操作，不是按钮
6. **挂载自检**：动作所在工作表必须是操作的**发起方**，而非目标表
7. **闭环自检**：`createRelatedRecord` 必须有对应的目标工作表，且**必须确保源工作表与目标工作表之间已经显式建立并声明了关联关系（Relation 字段）**。大模型必须在动作的 `relateFieldName` 中指明具体的关联字段名称，且该字段必须在源工作表的 `fields` 列表中显式存在。如果源表和目标表之间在规划中没有直接的关联字段，则禁止将其规划为 `createRelatedRecord`，而应当：
   - (1) 在源工作表字段列表中显式追加并声明 Relation 关联字段，并在 `relateFieldName` 中引用它；
   - (2) 或者将自定义动作的类型设计为 `triggerWorkflow`（通过后台工作流在点击时自动创建对应记录并建立数据同步）。
8. **字段修改自检**：`updateCurrentRecord` 必须有实际需要填写的字段。

## 四、`customPages[].components`（紧凑字符串数组）

1. 必须指定 `pageType`：`"dashboard"`（数据统计）或 `"workspace"`（工作台）
2. `description`：一句话说明该页面的业务目的与目标用户（如「面向管理层的借阅数据全局统计看板」）
3. **图标固定**：`dashboard` 固定使用 `"sys_control-panel_traffic"`，`workspace` 固定使用 `"2_3_statistics"`。无需从 icon guide 挑选
4. 组件格式：`"组件名(Type)"`
5. **命名规范**：组件名称必须具体反映业务含义，严禁使用"按钮1"、"业务分析"等模糊命名

**dashboard 页面要求：**
- 固定包含 4 或 6 个 `NumberChart`，2～6 个业务图表，1～2 个 `PivotTable`
- 每个图表标注数据源：`"组件名(ChartType:工作表名)"`
- 禁止用维度表（分类表）作数据源，必须用事实表（订单表）

**workspace 页面要求：**
- 主要是 `Button`（4～6 个）和 `View`（2～4 个）
- 可辅以 `Text`、`Carousel`、`Section`
- 展示型推荐内容（如新书推荐、热门商品、精选案例）应使用 `Carousel` 而非 `View`。

## 五、`workflows`

1. `description`：一句话说明该工作流的业务目标（如「借阅到期前 1 天自动提醒借阅人归还」）
2. `trigger.type`：`worksheet_event` / `schedule` / `date_field`
2. `trigger.label`：触发节点展示文字，如「借阅记录新增时」
3. `trigger.source`：`worksheet_event` → 工作表名称；`date_field` → `工作表名称（日期字段名）`；`schedule` → `""`
4. `intentHints` 是**业务效果与约束**数组，不要写成节点步骤：
   - ✅ `"通知内容应包含图书名称、作者、分类等关键信息"`
   - ❌ `"获取图书的书名、作者、分类信息"`
   - ✅ `"仅处理状态为借阅中的记录"`
   - ❌ `"查询所有借阅中的记录"`
   - **禁止引用具体字段选项值**（如「全部借出」），只描述业务目标
5. **不输出** `CustomAction` 触发类型的工作流到此处，它们已内嵌在 `worksheetCustomActions` 的 `intentHints` 中

## 六、`roles`

角色对象下字段：

- `name`：角色名称
- `description`：一句话说明该角色的职责定位与权限范围（如「负责日常借阅登记与归还操作的前台工作人员」）
- `roleScope`：角色类型。`"general"` = 组织内部角色（默认）；`"externalPortal"` = 外部门户角色
- `permissions`：紧凑字符串数组，格式 `"名称(类型)"`，如 `"图书(worksheet)"`、`"运营看板(customPage)"`

## 七、`aiAssistants`

直接复用总览「AI 助手」章节的内容；总览中没有时传空数组 `[]`。

1. `description`：一句话说明该助手的服务场景与能力范围（如「帮助读者查询图书库存、推荐书目、解答借阅规则」）
2. **图标固定使用 `"17_6_reddit"`**，无需从 icon guide 挑选

## 八、`navGroups`

1. 每个分组的 `items` 用紧凑格式：`"名称(worksheet)"`、`"名称(customPage)"`、`"名称(aiAssistant)"`
2. 所有工作表、仪表盘、AI 助手必须全部出现在某个分组中，**不遗漏**
3. 名称必须与方案中定义的名称**完全一致**

---

## ⚠️ 提交前闭环自检

写入 `hap-plan.json` 前必须检查：
1. 每个视图筛选条件是否有对应的字段选项值
2. 每个时效性状态选项是否有配套的自动标记工作流
3. 每个自定义动作的启用条件是否完整
4. 每张表的 `consistencyNotes` 是否覆盖了视图筛选、工作流读写的关联说明
5. `navGroups` 是否包含了所有工作表、自定义页面和 AI 助手
6. **字段丰富度**：每张主数据表是否 ≥ 15 个字段、业务单据表 ≥ 12 个、过程表 ≥ 8 个
7. **7 维度覆盖**：每张表是否覆盖了主标识、核心业务属性、状态分类、时间、责任协作、数量金额、说明凭证 7 个维度
