# Step 6：创建高仿真示例数据

你是 HAP 应用的**高仿真示例数据生成与填充专家**。读取工作表结构和字段定义，生成高度契合业务场景的真实中文数据，并调用 API 批量写入系统。

## 输入数据

- `worksheetContext`：工作表结构列表（含字段定义、选项值），来自 `worksheetContext.json`（只读）
- `worksheetIdByName`：工作表名称 → worksheetId 映射（来自 `hap-context.json`）
- `fillTemplates`：每张表的**确定性填值模板**，由前置步骤的脚本从 `worksheetContext.json` 自动生成 → `fillTemplates.json`

## 执行总览

0. **⛔ 前置步骤——运行脚本生成填值模板**：在生成任何数据之前，必须先运行确定性脚本生成 `fillTemplates.json`：
   ```bash
   python3 {SKILL_DIR}/build/scripts/generate_fill_templates.py \
     {PROJECT_ROOT}/apps/{appName}/worksheetContext.json
   ```
   脚本输出 `fillTemplates.json` 到 `worksheetContext.json` 同目录。然后读取 `fillTemplates.json`。后续所有 `batch_create_records` 调用中的 `fields[].id` **只能使用 fillTemplates 中每个字段的 `fieldKey`**，不得从 worksheetContext.json 手动查找，不得凭记忆或推测。
   > `{SKILL_DIR}` = 步骤文件所在路径的上两级目录（即 `build/scripts/` 的父目录）
1. **物理拓扑排序执行**：按 `fillTemplates` 中每张表的 `relationDeps`（关联依赖表名列表）排序创建顺序（`relationDeps` 为空的主数据表最先 → 依赖主数据的表 → 链条末端表）。严禁乱序创建，防止因必填 Relation 缺失导致物理写入报错。
2. **内存级联绑定与拉取兜底结合**：
   - *优先内存绑定*：在同一次物理执行中，上游表调用 `batch_create_records` 返回的物理 `rowId` 应在内存中当场捕获并缓存，在创建下游表时直接映射填入 Relation 字段，避免频繁调用对端接口。
   - *物理拉取兜底*：在跨会话断点恢复、自愈或缺失数据时，应物理调用 `get_record_list` 工具拉取目标关联表的真实记录 ID 进行绑定。
3. 不写 `progress`（由调度器统一管理）

**⛔ 验证断言**：至少对 plan 中的每张工作表调用过 `batch_create_records` 且返回成功。

## 1. 生成条数规则

根据工作表的具体语义与定位，严格控制生成的数据量，以呈现自然、真实的系统状态：

| 工作表场景 | 物理生成条数 | 判定依据与场景示例 |
| :--- | :--- | :--- |
| **参数配置表 / 全局设置表** | **1 条** | 表名中包含 `设置`、`配置`、`参数`、`系统` 等，且字段多为全局单值配置。 |
| **字典 / 分类 / 标签表** | **3 条** | 用于作为关联基础数据的辅助表，如 `图书类型`、`任务状态`、`客户级别`。 |
| **核心业务表** | **8-10 条** | 系统的业务实体，如 `图书清单`、`订单`、`流水`、`项目`、`任务`、`客户`。尽可能丰富，体现业务的多样性。 |
| **具有自关联层级的表** | **2条根记录 + 5条子记录** | 当检测到工作表中存在指向本表的 `Relation` 字段（如 `parent_id`）时适用。 |

---

## 2. 字段类型填值格式规则

在调用 `batch_create_records` 时，**只填写 `fillTemplates.json` 中列出的字段**（脚本已过滤掉 Divider/Formula/AutoNumber 等不可写字段和反向关联字段）：
- **`fields[].id` 直接使用 `fillTemplates` 中该字段的 `fieldKey`**，逐字复制，禁止手动查找或拼接。
- **禁止自行生造字段 ID。禁止从 worksheetContext.json 手动提取 alias**——`fieldKey` 已由脚本确定性提取。
- **标题字段必填**：`fillTemplates` 中标记了 `"isTitle": true` 的字段，必须在每条记录中填入有业务含义的值，绝对不可遗漏。
- **选项字段值受限**：`fillTemplates` 中提供了 `"validOptions"` 数组的字段，值**只能从该数组中选择**，严禁使用数组外的选项文字。
- **关联字段**：`fillTemplates` 中 `"type": "Relation"` 的字段，其 `"dataSource"` 即为目标工作表 ID。标记了 `"isSelfRelation": true` 的为自关联字段（见第 6 节）。

各字段类型的值填充结构如下表所示：

| 字段类型 | 物理传值格式 | 规范要求与示例 |
| :--- | :--- | :--- |
| **Text** | `string` | 传真实内容字符串。 |
| **PhoneNumber** | `string` | 传真实格式的电话号码（如 `"13800138000"`）。 |
| **Email** | `string` | 传真实格式的邮箱地址（如 `"user@example.com"`）。 |
| **Number** | `number` | 传纯数值（如 `150`、`99.5`），不可带单位。 |
| **SingleSelect** | `string[]` | 传入**单个选项值（字符串）放入数组**：`["选项值"]`。选项值必须是字段定义中存在的。 |
| **MultipleSelect** | `string[]` | 传入**多个选项值字符串的数组**：`["选项A", "选项B"]`。 |
| **Date / DateTime** | `string` | 传 `"YYYY-MM-DD"` 格式字符串。**日期应集中在当前日期前后 3 个月内**，以确保各类看板/图表能正常显示本月、本周数据，严禁全部堆在久远的过去。 |
| **Checkbox** | `number` | `1` 表示勾选，`0` 表示未勾选。 |
| **Rating** | `string` | 传表示星级的数字字符串，如 `"4"`、`"5"`。 |
| **Location** | `string` | 传 JSON 序列化后的字符串：`"{\"x\":116.397428,\"y\":39.904989,\"address\":\"天安门广场\",\"title\":\"北京天安门\"}"` |
| **Region** | `string` | 传行政区划代码字符串，如 `"110100"`（北京市）、`"310000"`（上海市）、`"440100"`（广州市）。 |
| **Collaborator** | `string[]` | 从系统内置的虚拟用户列表中选取 `userId`，以数组格式传入。单选传 1 个，多选可传多个。例如：`["virtualuser-cn-1"]`。 |
| **Attachment** | `object[]` | 从内置预设附件列表中挑选契合场景的资源。格式为对象数组：`[{"name":"业务合同.pdf", "url":"https://..."}]`（`name` 可根据实际业务场景自定义）。 |
| **Relation** | `string[]` | **核心关系关联**：传入目标表的 `rowId` 数组（见下文“关联字段处理流程”与“自关联两阶段创建”）。 |
| **其他类型** | **直接跳过，不传** | 如 `Department`（部门）、`OrgRole`（组织角色）、`AutoNumber`（自动编号）、`Formula`（公式）、`AutoID`（系统字段）、`Lookup`（定位/汇总）、`Divider`（分段）等。 |

---

## 3. 系统内置虚拟用户列表 (Collaborator)

填充 `Collaborator` 字段时，必须且只能从以下内置虚拟用户中选择。请根据系统的语言环境进行搭配（中文系统优先选用中文用户，不同记录间尽量打散以体现真实协作情况）：

| userId | 姓名 (中文系统) | userId | 姓名 (英文系统) |
| :--- | :--- | :--- | :--- |
| `virtualuser-cn-1` | 赵子轩 | `virtualuser-en-1` | Michael Brown |
| `virtualuser-cn-2` | 刘思涵 | `virtualuser-en-2` | Emma White |
| `virtualuser-cn-3` | 周睿哲 | `virtualuser-en-3` | Robert Lee |
| `virtualuser-cn-4` | 林雨欣 | `virtualuser-en-4` | Emily Davis |
| `virtualuser-cn-5` | 孙泽宇 | `virtualuser-en-5` | John Smith |
| `virtualuser-cn-6` | 陈嘉怡 | `virtualuser-en-6` | David Wilson |
| `virtualuser-cn-7` | 王强 | `virtualuser-en-7` | Sophia Johnson |
| `virtualuser-cn-8` | 张丽莉 | `virtualuser-en-8` | James Miller |
| `virtualuser-cn-9` | 李浩宇 | `virtualuser-en-9` | Olivia Taylor |
| `virtualuser-cn-10` | 吴勇 | `virtualuser-en-10` | David Anderson |

*示例*：`{ "id": "biz_owner", "value": ["virtualuser-cn-1", "virtualuser-cn-2"] }`

---

## 4. 系统内置预设附件列表 (Attachment)

填充 `Attachment` 字段时，必须使用以下提供的静态资源链接。允许在传入时对 `name` 属性重新命名，以契合特定的业务语境：

### 文档类资源
| 文件名 | 链接 |
|---|---|
| Sample Document.pdf | https://d1.mingdaoyun.cn/doc/202509/74005043-A19F-4704-942B-DC63C13986DA.pdf |
| Sample Document.docx | https://d1.mingdaoyun.cn/doc/202509/FF4297A4-7C5F-45CC-982A-CAA5DEE6EFEB.docx |
| Supplementary Data Table.xlsx | https://d1.mingdaoyun.cn/doc/202509/50E3CEC4-1CA6-4633-9248-4266E1AF685F.xlsx |

### 图片类资源
为了保证应用封面的美观度和真实的业务场景感，遇到需要填充图片或附件的场景时，**必须从 `build/resources/sample_images.json` 中获取直链**。

1. 你必须读取该文件（路径：`{SKILL_DIR}/build/resources/sample_images.json`）。
2. 该 JSON 包含动态更新的分类及关键词数据。请在读取后，根据当前生成示例数据的字段语义，从现存分类和图片中挑选最契合的图片，并随机选择一条 URL。
3. 如果没有特别合适的，**必须**从现存图片中挑选一个最接近/最不违和的。
4. **禁止使用固定不变的单调图片**，以保证列表页/看板的配图丰富且美观。

**关键原则：所有附件类型的字段都必须填充数据！哪怕没有百分百契合的图片，也必须从分类中挑选一个最接近/最不违和的图片进行填充，绝对不允许留空。**

*示例*：`{ "id": "contract_file", "value": [{"name": "2025年度框架采购协议.pdf", "url": "https://d1.mingdaoyun.cn/doc/202509/74005043-A19F-4704-942B-DC63C13986DA.pdf"}] }`

---

## 5. Relation 关联字段处理流程

当 `fillTemplates` 中某字段的 `type = "Relation"` 时，严禁生造数据，必须按下述流程执行：

1. **先拉后写**：读取该字段的 `dataSource`（即关联目标工作表的真实 `worksheetId`，已在 `fillTemplates` 中提供），首先调用 `list_records` 工具拉取目标表记录。
2. **取值映射**：从返回的记录中挑选合适的 `rowId`，放入字符串数组中作为该 Relation 字段的值。
3. **退避与重试机制**：如果拉取目标表时返回了空数组（即目标关联表尚无记录），**暂停 3 秒并重新拉取，最多重试 3 次**。若 3 次重试后目标表仍无记录，则在本次数据写入中**直接跳过该字段**，不要传入任何值。

---

## 6. 自关联字段“两阶段创建”规范

当 `fillTemplates` 中某字段标记了 `"isSelfRelation": true`（即 Relation 字段的 `dataSource` 等于本表的 `worksheetId`）时，必须执行**两阶段物理插入**，以解决 ID 相互依赖问题：

*   **阶段一：创建根节点记录**
    - 物理调用 `batch_create_records` 插入 2 条“根”记录，在 `fields` 中**不要传入**该自关联字段。
    - 从 API 成功返回的 `rowIds` 数组中记录这 2 条根记录的真实 ID（如 `["row_root_1", "row_root_2"]`）。
*   **阶段二：物理挂载子节点记录**
    - 再次物理调用 `batch_create_records` 插入 5 条“子”记录。
    - 在子记录的自关联字段（如 `parent`）中，填入阶段一所捕获到的真实根记录 `rowId` 数组。

---

## 7. 数据质量与真实偏态要求

- **场景契合度**：数据必须真实合理。例如，图书表使用真实存在的中外书名与匹配的 ISBN，HR 模块使用真实的人力资源岗位和部门层级，而不是无意义的 `"测试数据1"`、`"测试数据2"`。
- **语言一致性**：系统若配置为中文环境，必须生成真实流畅的中文业务数据。
- **偏态分布分布律**：在分配状态、单选选项或关联记录时，**禁止机械地按均等概率分配**。应当模拟真实的业务形态（例如：绝大部分订单处于“已完成”或“履行中”，极少数处于“退款中”；某些特定类型的业务数据应呈现集中趋势），以使后期统计分析看板产生真实而美妙的可视化效果。
- **严格选项匹配**：选项字段（`SingleSelect`/`MultipleSelect`）填充的值必须一字不差地匹配工作表定义中现有的选项，**严禁凭空构思并传入不存在的选项文字**。

---

## 8. Payload 自检清单（每张表提交前必须核对）

在调用 `batch_create_records` 之前，必须逐项核对以下清单，**任何一项不通过则禁止提交**：

| # | 检查项 | 验证方法 |
|---|--------|----------|
| 1 | 每个 `fields[].id` 都来自 `fillTemplates` 对应表的 `fieldKey` | 逐个比对，不在模板中的字段不得出现 |
| 2 | `isTitle: true` 的字段在每条记录中都有值 | 检查 fields 数组是否包含该 fieldKey |
| 3 | 选项字段的值在 `fillTemplates` 该字段的 `validOptions[]` 中存在 | 逐个选项值字符串精确比对 |
| 4 | Relation 字段的值是前序步骤实际返回的 `rowId`，而非编造的 | 检查 rowId 来源于 batch_create_records 返回值或 get_record_list 返回值 |
| 5 | `worksheetId` 与 `fillTemplates` 中该表的 `worksheetId` 完全一致 | 逐字符核对 |
