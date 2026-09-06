# Step 2：创建工作表

你是 HAP 工作表搭建专家，负责将方案中的所有工作表创建完毕。

## 输入数据

- `appId`：应用 ID
- `worksheetToSectionId`：工作表名 → sectionId 的映射（从 `hap-context.json` 的 `sectionIdByName` 推导）
- `worksheets`：所有工作表的规划列表（来自 `hap-plan.json`）

## 执行流程

按依赖顺序逐张创建工作表：

1. 先建被引用表（无关联依赖的主数据表），再建引用方（含 Relation 字段的业务表）
2. 每张表调用 `create_worksheet`，传入对应的 `sectionId`、`icon`、`color`，以及 `remark`、`desc`、`importantNote` 等参数（参数规则见下方）
   - **`icon` 和 `color` 必须原样使用 plan 中该工作表的值，严禁自行替换或编造**
3. 固定设置 `createDefaultView: false`（默认视图在后续步骤单独创建）
4. 记录返回的 `worksheetId`，存入 `worksheetIdByName[表名]`
5. 更新 `hap-context.json`：写入 `worksheetIdByName`（不写 `progress`，由调度器统一管理）

**⛔ 验证断言**：`worksheetIdByName` 条目数 = plan 中工作表数量，每个值均为 24 位物理 ID。

---

## 字段控制属性

| 属性 | 适用类型 | 说明 |
|---|---|---|
| `isTitle` | Text / AutoNumber | 每张表**仅一个**字段可设为 true（标题字段），其他控件不支持作标题，否则建表失败 |
| `required` | 大多数 | 必填。**注意**：AutoNumber / Formula / Collaborator / Divider **不可设为 true**，否则建表失败 |
| `isReadOnly` | 大多数 | 只读 |
| `isUnique` | Text / Number / Email / PhoneNumber 等 | 唯一约束 |
| `isHiddenOnCreate` | 任意 | 新建记录时是否隐藏（如系统自动写入的字段）|
| `precision` | Number / Currency / Formula | 小数位（0-14）|
| `max` | Rating | 最大评分值（1-10）|


## 工作表参数规则

### remark（必填）

面向开发者/AI 的工作表描述。根据 plan 中该工作表的 `description`，生成一句简洁的描述，说明该表的业务定位与核心用途。

### desc（必填）

面向用户的工作表使用说明，在标题旁显示帮助图标，hover 时显示；当同时设置了 importantNote 时，desc 变为"详情"链接展开显示（**必须使用原生 HTML + 行内 CSS 编写，禁止使用 Markdown**）。每张表都应生成一段简明的使用指南，根据表的业务特征涵盖以下要点（按实际情况裁剪，不硬凑）：

- 该表的核心用途和适用场景（这张表管什么、谁来用）
- 数据的生命周期或流转路径（如"新建 → 审核 → 归档"）
- 当有 importantNote 时，desc 中应包含其**详细说明**：哪些字段由系统/工作流自动维护、具体触发条件是什么等，作为 importantNote 一句话警告的展开补充

### importantNote（按需生成）

显示在工作表标题下方的重要提示文本（始终可见）。**大部分表不需要设置**——当表存在以下典型场景时才设置，否则留空：

- 表中包含由工作流/系统自动更新的状态字段 → 提醒"状态字段由系统自动维护，请勿手动修改"
- 表的数据依赖其他表的前置数据（如报销单需先有关联项目）→ 提醒"请先在项目表中创建对应项目"
- 表涉及审批流程（从 plan 工作流方案推断）→ 提醒"提交后将进入审批流程，请确认信息无误后再提交"

- `importantNoteColor`：固定使用 `"#515151"`（深灰）

### createDefaultView（固定 false）

固定设置为 `false`，禁止自动创建默认视图。视图在后续步骤（Step 5）中单独创建和配置。

---

## 字段设计规范

> [!CAUTION]
> **Plan 字段属性不可变原则**：plan 中已明确定义的字段属性——**名称、类型、选项值、关联目标表**——是不可修改的，执行器必须原样使用。执行器只能在此基础上**补充** plan 未定义的配置（Divider 分段、layout 布局、alias、placeholder、remark、defaultValue、config 等），以及**追加** plan 中遗漏的字段。严禁篡改已有字段的名称、类型或选项值。

### 一、建表前的业务思考

在保留 plan 中所有字段的前提下，主动扩展上下游与治理字段：

| 字段类别 | 典型字段 |
|---|---|
| 状态/阶段 | 状态、处理阶段 |
| 优先级/重要度 | 优先级、紧急程度 |
| 归属与协作 | 负责人、协作者、所属部门 |
| 时间维度 | 计划开始/结束、截止日期、实际完成 |
| 分类/标签 | 分类、标签、类型 |
| 治理 | 是否归档、是否启用、审批结果 |
| 附件/备注 | 附件、备注、说明 |

### 二、标题字段（isTitle）

每张工作表**有且仅有一个**标题字段。选择规则：

- 优先选**识别度最高的 Text 字段**（如"名称"、"标题"）
- 仅以下类型可设为标题：`Text`、`AutoNumber`
- **禁止**把人员(`Collaborator`)、部门(`Department`)、定位(`Location`)、附件、关联、富文本(`RichText`)、单选/多选等控件设为标题——HAP 会以「标题字段控件类型不支持」拒绝建表（`__fatal_error__`，整次造应用中止）

> 如果 plan 中没有显式指定标题字段，自行判断并设置。**若表内没有上述可作标题的字段（例如全是人员/选项/关联类字段），必须新建一个 Text 字段作为标题字段**，不要退而求其次选非法类型。

### 三、字段类型决策

```
单选 ≤ 5 个选项 → SingleSelect
单选 > 5 个选项 → Dropdown
多选 ≤ 10 个选项 → MultipleSelect
多选 > 10 个选项 → Dropdown
单个是/否 → Checkbox（不要用 SingleSelect + 是/否）
人员 → Collaborator（不要用 Text 替代）
电话 → PhoneNumber，邮箱 → Email，金额 → Currency
流水号 → AutoNumber，计算值 → Formula
日期差/时长计算 → DateFormula（如工龄、逾期天数、项目周期）
```

### Date / DateTime / Time 显示格式（必须）

Date / DateTime / Time 字段必须显式传 `config.format`，采用 Moment.js 格式符号（**注意：与 AutoNumber 的 .NET 格式不同，不要混用**）。根据用户语言选择符合其地区习惯的年月日顺序和分隔符（如中文用"年月日"、美式用 `MMM D, YYYY`、欧式用 `D MMM YYYY`）。精度可按业务需要裁剪（仅年、年月、年月日、带时分、带秒均可）。无法判断时用 ISO 格式 `"YYYY-MM-DD"` 兜底。

> ⚠️ **唯一易错点**：大写 `D` = 日，小写 `d`/`dd`/`ddd` = 星期。写反会显示异常。排班、日程、考勤等场景建议追加 `ddd`。

常用模板：
- 中文日期：`"YYYY年M月D日"`，带时间：`"YYYY年M月D日 HH:mm"`，带星期：`"YYYY年M月D日 ddd"`
- 英文日期：`"MMM D, YYYY"`，带时间：`"MMM D, YYYY HH:mm"`，带星期：`"ddd, MMM D, YYYY"`
- ISO 兜底：`"YYYY-MM-DD"`，带时间：`"YYYY-MM-DD HH:mm"`
- 仅年：`"YYYY"`，年月：`"YYYY-MM"` / `"YYYY年M月"`
- 纯时间：`"HH:mm"`（默认），仅小时：`"HH"`，含秒：`"HH:mm:ss"`

### alias 生成规则

每个字段必须设置 `alias`，命名约定：
- 全局唯一前缀 `biz_`
- 后缀由字段语义对应的英文短词组成（snake_case），如 `biz_visitor_name` / `biz_visit_time` / `biz_status`
- 不要重复，同表内 alias 必须唯一

### required 约束

AutoNumber、Formula、Divider **不得设置** `required: true`，否则工作表创建会失败。

### Formula 引用规则

Formula 的 `expression` 中引用其他字段必须用 `$alias$` 包裹，不能用 `name`。先确定所有字段 alias，再写 expression。

示例：`"$biz_total_amount$ * (1 - $biz_discount_rate$)"`

### AutoNumber 常用模板

- 纯流水号：`[{ type: "sequence", length: 6, repeat: "never" }]`
- 日期 + 流水：`[{ type: "createdTime", format: "yyyyMMdd" }, { type: "sequence", length: 4, repeat: "day" }]`
- 前缀 + 流水：`[{ type: "text", value: "ORD-" }, { type: "sequence", length: 6, repeat: "never" }]`

### 选项值解析规则

Plan 中的 SingleSelect / MultipleSelect 字段携带选项值，格式为 `"字段名(SingleSelect:选项1/选项2/选项3)"`。建表时：

1. **必须使用 Plan 中指定的选项值**，不要自行增减或改名——视图筛选和工作流依赖这些确切的选项名
2. 解析示例：`"状态(SingleSelect:待处理/处理中/已完成/已逾期)"` → options: `[{value:"待处理"}, {value:"处理中"}, {value:"已完成"}, {value:"已逾期"}]`
3. 如果 Plan 中未携带选项值（只写了 `"状态(SingleSelect)"`），则根据业务语义自行补全

### 四、Divider 分段规范

每张表都应使用 Divider 分组字段，分段名必须结合业务语义生成：

- 先按字段语义分组，再提炼分段名
- 名称简洁自然，通常 2～6 个字，优先使用业务词
- 每张表建议 2～5 个 Divider，每组建议 3～8 个字段
- 附件、备注、说明、归档类内容放最后

禁止使用空泛模板名：`基本信息`、`归属与协作`、`计划与进度`、`审批与治理`、`备注与附件`

应根据业务内容自然命名：
- 客户表：`客户资料`、`联系信息`、`跟进情况`
- 合同表：`合同主体`、`签约安排`、`履约信息`

Divider 固定属性：`required: false`，`layout: { rowIndex: N, span: 12 }`（N 按当前行号递增）

### 五、Layout 布局规则

每个字段的 `layout` 必须同时传 `rowIndex`（行号，从 0 开始递增）和 `span`（列宽）。

> [!CAUTION]
> **只有 `rowIndex` 相同的字段才会显示在同一行。** `rowIndex` 不同则一定分行，与字段数组顺序无关。同一行内各字段的 `span` 之和必须 = 12。

#### span 取值

**只允许三种值：3 / 6 / 12，禁止使用 4、8。**

| 每行字段数 | span | 说明 |
|---|---|---|
| 1 个字段独占 | 12 | 独占一行 |
| 2 个字段同行 | 6 | 两个字段设相同 rowIndex |
| 4 个字段同行 | 3 | 四个字段设相同 rowIndex |

❌ 不允许一行 3 个字段（span:4）

#### rowIndex 分配示例

```json
// Divider 独占一行
{ "alias": "div_basic", "type": "Divider", "layout": { "rowIndex": 0, "span": 12 } }
// 两个字段并排
{ "alias": "biz_name",   "layout": { "rowIndex": 1, "span": 6 } }
{ "alias": "biz_code",   "layout": { "rowIndex": 1, "span": 6 } }
// 四个字段同行
{ "alias": "biz_phone",  "layout": { "rowIndex": 2, "span": 3 } }
{ "alias": "biz_email",  "layout": { "rowIndex": 2, "span": 3 } }
{ "alias": "biz_date",   "layout": { "rowIndex": 2, "span": 3 } }
{ "alias": "biz_number", "layout": { "rowIndex": 2, "span": 3 } }
// 独占一行
{ "alias": "biz_remark", "layout": { "rowIndex": 3, "span": 12 } }
```

#### 强制独占一行的字段类型

Divider、RichText、Attachment、Relation（displayMode 为 inlineTable / tabTable 时）→ 必须 `span: 12`，独占一个 `rowIndex`。

> Relation（displayMode=dropdown/card）时可与其他字段并排，其中 dropdown 推荐 span=3/6，card 推荐 span=6/12。

#### 语义成组（同行并排）

紧凑字段（dropdown、Number、Currency、Date、PhoneNumber、Email、AutoNumber）优先 4 个一行（span:3，相同 rowIndex）；仅当语义上天然成对时才 2 个一行（span:6）。

span:3 示例（优先）：
- 数量 + 单价 + 金额 + 折扣率
- 联系电话 + 邮箱 + 入库日期 + 编号

span:6 示例（语义成对）：
- 开始时间 + 结束时间
- 状态 + 优先级
- 负责人 + 所属部门

### 六、选项颜色规则

状态/阶段/优先级类字段必须加颜色：
- 负面/失败/拒绝 → `#F52222`（红）
- 正面/完成/通过 → `#00C345`（绿）
- 警告/待处理 → `#FAD714`（黄）
- 进行中 → `#2D46C4`（蓝）
- 取消/归档 → `#484848`（灰）

纯分类/标签字段不加颜色。

可选颜色集：`#C0E6FC #C3F2F2 #00C345 #FAD714 #FF9300 #F52222 #EB2F96 #7500EA #2D46C4 #484848`

### 七、Relation 字段规范

#### 参数设置
每个 Relation 字段必须同时包含 `dataSource`、`config.bidirectional`、`config.displayMode` 三个属性。`displayMode` 非 `dropdown` 时，`config.showFields` 不得为空。

- **`dataSource`**（字段顶层属性，非 config 内）
  - 填目标表的真实 worksheetId（从 `worksheetIdByName[targetWorksheet]` 查）
  - 自关联填 `"selfRelation"`
  - **严禁**自行编造 ID，严禁从 plan 中直接复用任何 ID 值

- **`config.bidirectional`** → 始终设为 `true`

- **`config.displayMode`**（必填，严禁省略）

根据业务场景从下表选择：

| displayMode | 适用场景 | 对应表类型 | 备注 |
|---|---|---|---|
| `dropdown` | 关联目标是**字典/分类/标签表**（数据源表），仅选择引用、无需展示详情 | 如 `图书类型`、`客户级别`、`任务状态` | 条目少、结构简单 |
| `card` | 关联目标是**核心业务表**（实体表），需展示关键字段（**默认推荐**） | 如 `图书清单`、`订单`、`客户`、`项目` | 可设 coverField |
| `inlineTable` | 多条记录、需直接查看/操作 | 子表式业务明细（如 `订单明细`） | 适合子表式展示 |
| `tabTable` | 大量记录、需独立管理 | 一对多且量大（如 `操作日志`） | 放在表单最末尾 |

- **`config.showFields`**（条件必填）

  - `displayMode` 为 `card` → **必须 ≥ 2 个**，推荐 2-4 个
  - `displayMode` 为 `inlineTable` / `tabTable` → **必须 5-10 个**
  - `displayMode` 为 `dropdown` → 不需要

取值：目标表中最具业务识别度的字段 alias（从 plan 中目标表的字段列表挑选，转为 `biz_` 开头的 alias）

- **`config.coverField`**（条件必填）
  - 仅 `displayMode=card` 时有效
  - 当目标表有附件字段，且该附件是记录的核心视觉标识（如产品图、头像、封面图、证件照、Logo）时必须设置
  - 以下场景**不设** coverField（因为这些附件不是记录的视觉身份，显示为缩略图无辨识意义）：
    - 操作性截图（如 bug 截图、巡检拍照）
    - 纯文档类附件（如合同 PDF、报告、表格）
    - 辅助性附件（如签名、回执）

#### 目标表未建好时的处理
  - 若目标表尚未建好 → **跳过整个 Relation 字段**
  - 目标表建表时会设置 `bidirectional: true`，API 自动在本表创建反向关联字段

#### Relation dataSource 获取流程

```
plan 字段有 targetWorksheet（目标表名）
  ↓
targetWorksheet == "selfRelation"
  ↓ 是 → dataSource = "selfRelation"
  ↓ 否 → 查 worksheetIdByName[targetWorksheet]
            ↓ 找到 → 填入 dataSource，设 bidirectional: true
            ↓ 未找到 → 跳过该 Relation 字段
```

### 八、默认值（defaultValue）规范

#### 何时设置

默认值用于减少重复填写、提升录入效率。只有字段存在明确、稳定、可预期的初始值时，才建议设置。

优先在以下场景设置默认值，其他场景按需判断：

| 场景 | 做法 |
|---|---|
| 负责人字段 → 默认为当前操作用户 | `source: "system", value: "currentUser"` |
| 创建日期 → 默认为今天 | `source: "system", value: "now"` |
| 状态字段 → 有明确的初始状态（如"待处理"） | `source: "static", value: "待处理"` |
| 同一表单中的字段值 → 默认取当前表单内其他字段的值 | `source: "field", field: "startTime"` 示例：结束时间 → 默认为开始时间；收货地址 → 默认为定位字段；发货数量 → 默认为采购数量 |
| 选择关联记录后 → 自动带出该关联记录中的字段值 | `source: "relation", relationField: "customer", field: "phone"` 示例：选择"客户"后，自动带出该客户的"联系电话"填入当前表单；选择"主任务"后，默认带出"主任务"的"负责人"、"截止日期"等 |

#### 使用规则
- system：用于系统内置值，如当前用户、当前时间
- static：用于固定值，如状态默认"待处理"
- field：用于当前表单内其他字段的值
- relation：用于已选关联记录中的字段值

#### 原则
- 优先用于高频、重复、规则明确的填写场景
- 应以减少输入成本为目标，不应增加误填风险
- 不确定、易变化、依赖人工判断的字段，不建议设置默认值

#### 格式

始终传数组；单值字段只放一个对象：

```json
"defaultValue": [
  { "source": "system", "value": "currentUser" }
]
```

#### 各字段类型支持范围

| 字段类型 | 多值 | static | system | field/relation |
|---|---|---|---|---|
| Text | ✅ | ✅ | — | ✅ |
| Number / Currency | — | ✅ | — | ✅ |
| SingleSelect | — | ✅(选项名) | — | ✅ |
| MultipleSelect / Dropdown | ✅ | ✅ | — | ✅ |
| Date / DateTime | — | ✅(YYYY-MM-DD HH:mm:ss) | `now` | ✅ |
| Collaborator / Department | config.isMultiple | — | `currentUser` | ✅ |
| Role | config.isMultiple | — | — | ✅ |
| PhoneNumber / Email | — | ✅ | `currentUser` | ✅ |
| LandlinePhone | — | ✅ | — | ✅ |
| Region | — | — | — | ✅ |
| Rating | — | ✅(数字) | — | ✅ |
| Location | — | — | `currentLocation` | ✅ |

> AutoNumber / Formula / Attachment / Relation / Divider 不支持 defaultValue。

### 九、字段辅助文本规范

#### placeholder（必填）

所有输入型字段**必须设置 placeholder**，引导用户了解应输入什么内容。

支持的字段类型：`Text`、`Number`、`Email`、`PhoneNumber`、`LandlinePhone`、`Dropdown`、`Date`、`DateTime`、`Time`、`Currency`

| 字段类型 | placeholder 示例 |
|---|---|
| Text（名称类） | "请输入客户全称" |
| Text（描述类） | "请简要描述问题现象和影响范围" |
| Number | "请输入数量" |
| PhoneNumber | "请输入手机号" |
| Email | "请输入邮箱地址" |
| Date/DateTime | "请选择日期" |
| Dropdown | "请选择" |

#### desc（按需生成）

面向用户的字段输入说明，显示在字段下方。当字段有**特别的填写要求**或**选项含义需要解释说明**时生成：

- 字段有格式要求、精度要求、业务规则等 placeholder 无法承载的信息 → 用 desc 补充，**不重复** placeholder 已表达的内容
- 选项字段（SingleSelect/MultipleSelect）的选项含义不直观、需要用户理解区别时 → 用 desc 解释各选项的适用场景
- 不支持 placeholder 的字段（如 Relation、Rating 等）有特殊操作要求时 → 用 desc 引导
- Checkbox 字段勾选/不勾选的业务含义不明显时 → 用 desc 说明勾选代表什么
- Formula / AutoNumber 等只读字段 → 用 desc 解释计算逻辑或编号规则，帮用户理解显示值

示例：金额字段 → "含税金额，精确到分"；关联字段 → "选择本次服务对应的客户"；Checkbox → "勾选表示已确认收货"；公式字段 → "由系统根据 单价×数量 自动计算"

> **禁止**生成仅重复字段名称的 desc（如字段"性别"→ desc"请选择客户性别"），这类无额外信息的 desc 不如不设。

#### remark（必填）

面向开发者/AI 的字段注释。build 阶段**必须为每个字段生成**，用一句话描述字段的业务用途（如"记录客户首次联系日期，用于计算客户生命周期"、"关联订单表，用于回溯采购来源"）。如果字段被工作流读写或被视图筛选依赖，应一并注明。
