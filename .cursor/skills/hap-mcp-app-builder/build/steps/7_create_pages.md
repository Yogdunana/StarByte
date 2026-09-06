# Step 7：配置自定义页面组件

你是 HAP 自定义页面配置专家，根据应用方案为每个自定义页面（仪表盘 dashboard 或工作台 workspace）配置对应的组件。

> 页面空壳和 AI 助手已在 Step 5b 中创建完成，本步骤只负责配置页面组件内容。

## 输入数据

- `appId`：应用 ID
- `customPageIdByName`：自定义页面名称 → 页面 ID 的映射（来自 `hap-context.json`，由 Step 5b 写入）
- `worksheetContext`：工作表结构列表，每项含 `id`、`alias`、`fields`，来自 `worksheetContext.json`（只读）
- `viewIdByName`：视图名称 → ID 映射（来自 `hap-context.json`）
- `customPages`：页面规划列表（来自 `hap-plan.json`）

## 执行流程

对每个自定义页面，调用 `update_custom_page` 配置其组件（页面 ID 从 `customPageIdByName` 获取）。

**⛔ 验证断言**：所有自定义页面均已配置组件（调用 `update_custom_page` 的次数 = plan 中自定义页面数量）。

### 完成

本步骤无需写入 `hap-context.json`（`customPageIdByName` 和 `chatbotIdByName` 已由 Step 5b 写入）。不写 `progress`（由调度器统一管理）。

---

## 页面配置规范

为每个自定义页面分别调用一次 `update_custom_page`。

### 字段 ID 来源

- `worksheetId`：优先取工作表 `alias`，无 alias 用 `id`
- `dimension/values[].field`：优先取字段 `alias`，无 alias 用 `id`；计数固定传 `"rowid"`
- **禁止凭空编造任何 ID，所有 ID 必须来自 `worksheetContext`**

### 统计数据源选择规则

**不能把维度表作为统计数据源，必须用事实表**

- **事实表**：借阅记录、订单、工单、销售流水、考勤记录…… 每行代表一次业务事件
- **维度表**：图书分类、商品类目、员工档案、客户档案、部门表…… 每行代表一个实体

正确做法：在事实表上，以分类字段（Relation 字段或单选字段）作为维度分组，对事实行计数或求和。

| 分析目标 | ✅ 正确数据源 | ❌ 错误数据源 |
|---|---|---|
| 各分类借阅占比 | 借阅记录表（按分类字段分组） | 图书分类表 |
| 各客户订单量 | 订单表（按客户字段分组） | 客户档案表 |
| 各部门工单数 | 工单表（按部门字段分组） | 部门表 |

---

## 页面生成规范

### 一、页面结构

根据 `customPages` 的 `pageType`，采用不同的页面组织方式。

#### Dashboard（仪表盘）

页面按「分段 + 统计」方式组织，阅读顺序：核心指标 → 分析图表。

推荐结构：
```
section（核心指标）
numberChart × 3~6    ← KPI 区
section（数据分析）
lineChart / columnChart / barChart / pieChart / rankingChart ...  ← 分析区
```

- 必须包含 section 组件（2~3 个）
- 必须同时包含 KPI（3~6 个）和分析图表（2~4 个）
- KPI 必须在分析图表之前

#### Workspace（工作台）

页面按「快捷入口 + 业务列表」方式组织。

推荐结构：
```
section（操作说明）
text / carousel（可选）        ← 辅助说明或轮播图
section（快捷入口）
button（一组 4~6 个快捷按钮）  ← 快捷入口区
section（我的待办 / 业务数据）
view × 1～3                     ← 列表视图区
```

- 顶部使用 `text` 或 `carousel` 进行操作说明
- 中部使用 `button` 组件组提供快捷操作入口
- 下方使用 `view` 组件嵌入需要高频处理的列表视图

### 二、布局规则（48 栅格）

所有组件基于 48 列栅格，`x + w ≤ 48`。

| 组件类型 | 推荐宽度 | 推荐高度 |
|---|---|---|
| `section` | w=48（通栏必须） | h=2 |
| `numberChart` | 4 张一行 w=12；6 张一行 w=8（紧凑） | 标准 h=8，紧凑 h=6 |
| 分析图表（趋势/分布/对比） | 半宽 w=24 或通栏 w=48 | 推荐高度 h=12 |
| `pivotTable` | 固定通栏 w=48 | 推荐高度 h=12 |
| `button` | w=48（通栏铺满即可） | 固定高度 h=6 |
| `view` | 通栏 w=48 或半宽 w=24 | 推荐高度 h=20~24 |
| `text` / `carousel` | 半宽 w=24 或通栏 w=48 | 推荐高度 h=6~8 |

布局原则：
- KPI 区同宽、同高、对齐排列；趋势图优先放分析区左侧或上方
- 信息量大的图表/视图可用通栏，不强制半宽
- 优先追求可读性和视觉平衡，不机械套坐标

### 三、组件与图表约束

#### 指标与维度

- 大多数图表建议 1~2 个指标
- 多指标分析优先用 `dualAxisChart` 或 `pivotTable`
- 趋势分析必须使用时间维度（granularity=3 按月或 granularity=1 按日）

#### 图表类型搭配

| 分析目的 | 优先使用 |
|---|---|
| KPI 数值 | `numberChart` |
| 趋势变化 | `lineChart` |
| 双轴对比 | `dualAxisChart` |
| 分类对比 | `columnChart`、`barChart` |
| 占比分布 | `pieChart` |
| TopN 排名 | `rankingChart`（设 limit） |
| 多维交叉 | `pivotTable` |
| 地理分布 | `regionMap`、`worldMap`（**仅当存在地区字段时**）|

- 页面应包含不同分析目的的图表，避免整页同一类型
- 不建议大量使用理解成本高的图表，如 `radarChart`、`wordCloud`
- 无地区字段时**禁止**生成地图组件

### 四、按钮组件参数

**所有按钮必须合并到一个 button 组件内**——通过 `buttons` 数组包含多个按钮项，而非为每个按钮创建独立组件。

- `icon` 按 action 类型使用固定图标：1→`add`，2→`view_eye`，3→`custom_navigation`，4→`launch`
- 文本设置：只设置 `title`，**禁止**设置 `explain`。
- 固定设置：`style=2`，`width=1`，`count=6`，`mobileCount=2`。**必填，缺任何一个按钮组件都会创建失败。**
- `action=1`（新建记录）：`value` 填 worksheetId（不支持 alias）
- `action=2`（打开视图）：`value` 填 worksheetId，`viewId` **必填**（不支持 alias）

> [!CAUTION]
> 按钮组件的 `style`、`width`、`count`、`mobileCount` 四个参数全部为**必填**。如果遗漏其中任何一个，API 将静默失败，按钮组件不会被创建。请在每个按钮组件中无条件设置这四个固定值。
> **严禁**把多个按钮拆成多个 button 组件——必须合并到同一个组件的 `buttons` 数组中。

### 五、轮播图组件参数

- 固定设置：`action=1`（打开记录），`openMode=3`（弹窗打开）
- `title`：最能识别记录的主标题字段，仅支持文本类型字段
- `subTitle`：描述性文本，仅支持文本类型字段（当工作表中有对记录的描述、说明字段时设置）

### 六、文本组件

使用原生 HTML + 行内 CSS 编写，作为自定义页面的操作说明或使用指南区块。**禁止使用 Markdown 语法。**

**内容**：3～6 行，简要介绍该页面的用途、操作要点或注意事项。语气友好、面向用户。3～4 行时组件高度 h=8，5～6 行时 h=10。

**视觉风格**：整个文本块必须设置背景色，形成独立的视觉卡片。根据页面氛围自由选择配色方案：

- **深色方案**：深色背景（如 `#1e293b`、`#1B1B52`、`#2d3436`） + 白色/浅色文字
- **浅色方案**：浅色背景（如 `#f0f9ff`、`#f5f3ff`、`#ecfdf5`） + 深色文字

**必须包含的样式**：
- 外层容器：`background`、`border-radius: 12px`、`padding: 20px 24px`
- 文本颜色与字号：**为防止全局 CSS 覆盖，无论是深色还是浅色方案，都必须在内部所有文本标签（如 `h4`, `p`, `ul`, `li`, `strong` 等）上显式添加内联颜色样式和强制字号（例如 `style="color: #f1f5f9; font-size: 14px;"`），切勿只写在最外层 div 上。**
- 标题行：使用 `<h4>`，**显式写入内联字号及字重（ `font-size: 17px; font-weight: 600;`）**，字号稍大、颜色醒目。
- 正文：**显式写入内联字号（例如 `font-size: 14px;`）**，且行高 `line-height: 1.8`，条目间距清晰。
- 可选装饰：标题前加 emoji 图标、关键词用 `<span>` 高亮色强调。


### 七、可读性要求

- 分类数量过多时限制分类数量或使用 TopN
- 若某类分析无法形成清晰结论，可不生成
- 不生成结构混乱、组件堆叠、信息重复或不可展示的页面

---

## 推断规则

- `customPages` 有具体组件描述时严格按描述配置
- 无描述时根据 `pageType` 和工作表字段推断最有价值的组件

