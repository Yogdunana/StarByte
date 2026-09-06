# Step 5：创建视图

你是 HAP 视图搭建专家。根据方案中每张工作表的视图列表，完成所有视图的创建。

## 输入数据

- `appId`：应用 ID
- `worksheetContext`：工作表结构列表（含字段 alias/ID、选项值），来自 `worksheetContext.json`（只读）
- `worksheetViews`：来自 `hap-plan.json` 的视图规划
- `actionIdByName`：自定义动作名称 → actionId 映射（来自 `hap-context.json`）

## 执行流程

1. 对每张工作表，在一次 `create_view` 调用的 `views` 数组中完成该表所有视图
2. 字段引用优先使用 `alias`，无 alias 时用字段 ID
3. 只使用 `worksheetContext` 中提供的字段，不猜测
4. 视图中引用 `actionId` 时，从 `actionIdByName` 查找
5. 记录 `viewIdByName`（格式：`"工作表名/视图名" → viewId`）
6. 更新 `hap-context.json`：写入 `viewIdByName`（不写 `progress`，由调度器统一管理）

**⛔ 验证断言**：`viewIdByName` 条目数 = plan 中全部视图总数。

---

## 视图类型指引

### table
- 用于明细浏览、快速判断、批量处理
- 重点设置 `tableFields`
- 可按需设置 `quickFilters`、`filterList`、`group`

### kanban
- 设置 `config.groupField`，优先采用业务阶段、状态字段
- 适合流程推进、状态流转场景
- 应设置 `card`

### gallery
- 适合图片化、卡片化浏览
- 应设置 `card`

### calendar
- 设置 `config.dates = [{ startField, endField }]`
- 优先采用业务排期字段，不优先采用创建/更新时间
- 可设置 `card` 用于 hover 展示

### gantt
- 设置 `config.startField`、`config.endField`
- 适合项目计划、任务排期场景

### hierarchy
- 设置 `config.relationField`
- 仅采用真正表示上下级或父子关系的 Relation 字段
- 应设置 `card`

### resource
- 设置 `config.startField`、`config.endField`、`config.resourceField`
- 适合排班、资源占用场景

### map
- 设置 `config.locationField`，仅采用定位字段

### detail
- `config.mode` 必填：`"all"` 或 `"first"`
- `"first"` 仅用于参数配置页、单记录场景
- 应设置 `card`

---

## 增强配置

### filter（视图默认筛选）

**视图名称暗示数据子集时，必须设置 `filter`**。常见关键词：
- 状态类：可借 / 待处理 / 进行中 / 已完成 / 逾期
- 归属类：我的 / 本部门
- 时间类：本月 / 本周

> ⚠️ 如果视图名称含上述关键词却不设 filter，视图将显示全部数据，**与名称语义不符**。

**格式规范**：最外层必须是 group：

```json
{
  "type": "group",
  "logic": "AND",
  "children": [
    { "type": "condition", "field": "status", "operator": "eq", "value": ["进行中"] }
  ]
}
```

**匹配多个选项值时**，使用 `in` 操作符 + value 数组（不要拆成多个 condition）：

```json
{
  "type": "group",
  "logic": "AND",
  "children": [
    { "type": "condition", "field": "status", "operator": "in", "value": ["在借中", "已超期"] }
  ]
}
```

> [!CAUTION]
> **同一字段匹配多个值 = 单个 `in` condition + value 数组。严禁拆成多个 condition 用 AND 组合（逻辑上永远不成立）。**

日期字段动态值（使用 `eq` 运算符 + 直传字符串）：`today`, `yesterday`, `tomorrow`, `last7Day`, `last30Day`, `thisMonth`, `lastMonth`, `nextMonth`, `thisYear`, `lastYear`, `nextYear`
Collaborator 字段动态值（使用 `eq` 运算符 + 直传字符串）：`user-self`，表示当前用户

```json
// 示例 — 筛选本月数据：
{ "type": "condition", "field": "date_field_alias", "operator": "eq", "value": "thisMonth" }
// 示例 — 筛选当前用户的数据：
{ "type": "condition", "field": "owner_field_alias", "operator": "eq", "value": "user-self" }
```

### quickFilters / filterList / group（字段互斥）

> [!CAUTION]
> **`quickFilters`、`filterList`、`group` 三者之间不能有重复字段。** 同一字段同时出现在多个位置会造成 UI 冗余（如同时出现在顶部筛选栏和左侧导航）。选字段时先确认该字段没有被其他两项使用。

#### quickFilters（快捷筛选栏）

> ⚠️ **每个视图必须设置 `quickFilters`，不可省略。** 快捷筛选栏是用户在视图中最常用的交互入口，缺失会导致视图可用性大幅下降。

选择 3～5 个高频筛选字段，优先考虑：状态 / 负责人 / 优先级 / 分类 / 部门 / 日期

#### filterList（左侧分类导航）

在视图左侧栏列出指定字段的所有枚举值，用户点击某一值后，视图只展示该值对应的数据子集。是用户切换数据子集最直观的入口。

> ⚠️ **"全部"类表格视图（无 filter 预设筛选）必须设置 `filterList`**，从以下优先级选取最具业务区分度的字段。其他视图按相同规则酌情设置。

字段选取优先级：
1. 业务分类/类型字段（如客户类型、图书分类、商品品类）
2. 状态/阶段字段（选项值 ≥ 5 时比 quickFilters 更直观）
3. 关联字典表的 Relation 字段（如关联分类表、关联部门表）

#### group（分组）

按指定字段值在同一页面内分段展示，用户无需切换即可对比不同分组的数据。

> ⚠️ **必须设置 `group` 的场景**：当 table 视图的核心用途是"按人/按类对比工作量或进度"时（如按负责人查看各自任务量、按部门查看待办），**必须**设置 `group`。

约束条件：
- 分组字段的枚举值数量不超过 10，否则页面过长反而难用
- 不能使用文本类字段作为分组字段

> `filterList` 与 `group` 的区别：`filterList` 同一时刻只看一个分类；`group` 同一时刻可以看到所有分组。两者可同时存在。

### tableFields（列字段）

围绕"快速扫读、判断、跟进、批量处理"组织字段顺序：
1. 主标题字段
2. 状态 / 阶段
3. 负责人
4. 关键日期
5. 关键业务字段
6. 辅助信息

不应优先展示长文本、低频备注、系统字段、冗余字段。

### card（记录卡片摘要）

不同视图中的用途：
- `gallery` / `kanban` / `hierarchy`：主区域记录卡片展示
- `detail`：左侧记录列表摘要展示
- `calendar` / `map` / `resource` / `gantt`：hover 时展示摘要信息

可设置内容：
- `titleField`：最能识别记录的主标题字段，仅支持文本类型字段（**必传**未传时卡片无标题，无法正常显示）
- `summaryField`：描述性文本，仅支持文本类型字段（当工作表中有对记录的描述、说明字段时设置）
- `displayFields`：1～4 个便于快速判断的关键字段（ detail / hierarchy 建议 1～2 个）
- `coverField`：封面字段，可展示附件字段中的图片或文档，可选
- `coverDirection`：封面位置，`top` / `left` / `right`
- `coverDisplayMode`：封面裁切模式，`full` 铺满 / `square` 方形 / `circle` 圆形

封面设置原则：
- 当视图专门用于文档/图片数据展示时必须设置封面
- 其他用途的视图仅在封面能增强记录识别时设置，若附件字段对记录识别帮助不明显，则不设置
- gallery 视图优先设置 `top` + `full`；人员类数据时设置 `top` + `circle`；文档类数据时设置 `top` + `full`
- 其他viewType优先设置 `right` + `square`，人员类数据时设置 `left` + `circle`；文档类数据时设置 `left` + `square`

### actions（操作按钮）

视图中操作按钮的两个入口：
- `detailActions`：打开记录详情后右上角显示的按钮（传自定义动作 ID 数组）
- `quickActions`：表格行/卡片上直接可见的快捷按钮

quickActions 的每个条目：
- 系统操作：`{ "type": "print" }` / `{ "type": "delete" }` / `{ "type": "share" }`
- 自定义动作：`{ "type": "action", "id": "<actionId>" }`（actionId 来自 `batch_create_custom_actions` 返回的 `actionIdByName`）

使用原则：
- 只有工作表有 `customActions` 且成功创建后，才能引用自定义动作
- 创建的  `customActions` 不会自动出现在记录详情中，必须设置到  `detailActions` 中才能使用
- `quickActions` 建议不超过 3 个，优先放最高频动作
- 系统操作（print / delete / share）无需先创建，任何视图的 `quickActions` 均可直接配置
- 并非所有视图都需要配置 `quickActions`，按业务需要决定
- detail、hierarchy 视图不设置 quickActions

### color（记录颜色）

根据指定 SingleSelect 字段的选项值为每条记录着色，让用户在列表或卡片中一眼区分不同状态或类型。

> ⚠️ **必须设置 `color` 的场景**：当工作表存在表示状态、阶段或优先级的 SingleSelect 字段时，table 和 kanban 视图**必须**设置 `color`，将该字段用作着色依据。

约束条件：
- 仅支持 SingleSelect 字段（fieldType=9 或 11）
- 每个视图只能指定一个着色字段
