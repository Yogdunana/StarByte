---
name: build
description: HAP 应用物理搭建调度器。读取 hap-plan.json，逐步调度 steps/*.md 完成所有 HAP 对象的创建与配置。支持 subagent 委派和内联执行双模式。
---

# HAP 应用搭建调度器

你是 HAP（明道云）应用搭建**调度器**，只负责进度路由、上下文合并和结果校验，不直接承载各步骤的详细规则。

> 本文件是物理搭建阶段的轻量调度器，不直接包含各步骤的详细搭建规则。
> 具体规则必须从 `steps/` 中对应步骤文件读取。
> 无论使用 subagent 还是内联执行，每一步都必须遵守 `OUTPUT_CONTRACT.md`。

> 本 skill 由根 `SKILL.md` 路由调用，`appName` 和 `org_id` 已在前置检查阶段确定。

---

## 🔒 全局执行清单（标记 completed 前必须核对）

> [!CAUTION]
> 以下每一项都是**硬性交付物**。标记 `progress="completed"` 之前，必须逐项核对并确认全部完成。遗漏任何一项即为执行失败。

| # | 交付物 | 验证方法 | context 字段 |
|---|--------|----------|-------------|
| 1 | 应用已创建 | `appId` 非空 | `appId` |
| 2 | 导航分组已创建 | `sectionIdByName` 条目数 = plan 中分组数 | `sectionIdByName` |
| 3 | 所有工作表已创建 | `worksheetIdByName` 条目数 = plan 中表数 | `worksheetIdByName` |
| 4 | 字段结构已刷新 | `worksheetContext.json` 非空且每表有 fields | `worksheetContext.json` |
| 5 | 自定义动作已创建 | `actionIdByName` 条目数 = plan 中动作数 | `actionIdByName` |
| 6 | 视图已创建 | `viewIdByName` 条目数 = plan 中视图总数 | `viewIdByName` |
| 7 | 示例数据已写入 | 各表均有记录 | `progress >= sample_data_created` |
| 8 | 页面空壳已创建 | `customPageIdByName` 条目数匹配 plan | `customPageIdByName` |
| 9 | AI 助手已创建（若有） | `chatbotIdByName` 条目数匹配 plan（或 plan 无则跳过） | `chatbotIdByName` |
| 10 | 页面组件已配置 | 所有页面均已调用 `update_custom_page` | Step 7 完成 |
| 11 | 角色已创建 | `roleContext` 条目数 = plan 中角色数 | `roleContext` |
| 12 | 工作流已设计 | `hap-plan.json` 中每条 workflow 和 customActionWorkflow 的 `nodes[]` 非空 | `hap-plan.json` |
| 13 | 系统工作流已发布 | 每个系统工作流 processId 已 publish | Step 10 完成 |
| 14 | 自定义动作工作流已发布 | `customActionWorkflows[]` 每条均已 publish | Step 11 完成 |
| 15 | CLI 建后精修已对账 | CLI 可用→已校正组织+设当前应用+回填 `cliGaps[]`；hap 未安装→已输出待补清单且未中断 | Step 12 完成 |

---

## 断点恢复

1. 读取 `{PROJECT_ROOT}/apps/{appName}/hap-plan.json`，提取 `org_id` 和所有规划数据
2. 检查 `{PROJECT_ROOT}/apps/{appName}/hap-context.json`
   - 存在 → 读取 `progress` 字段，从断点继续
   - 不存在 → 从 Step 1 开始
3. 如果进度 >= `fields_refreshed`，读取 `{PROJECT_ROOT}/apps/{appName}/worksheetContext.json` 加载字段结构

> 详细的 context 结构见 `build/CONTEXT.md`，progress 状态定义见 `build/PROGRESS.md`。

---

## CLI 自检（非阻断）

> 本步在断点恢复读出 `org_id` 之后、进入执行循环之前运行一次。
> 目的：把本机 `hap` 命令行工具**准备到「已安装且已登录」**，供最后的 Step 12 用 CLI
> 回填 MCP 盖不到的硬缺口。
>
> [!IMPORTANT]
> **本步不阻断搭建主体**——纯 MCP 已能独立把应用建好。**唯一会让 Step 12 跳过回填的情况是
> 「hap 未安装」**；「未登录」在本步自动登录解决，「组织不一致」留到 Step 12 自动切换解决。

1. **探测 hap 是否安装**：运行 `hap auth whoami`
   - **命令不存在 / 未安装** → `cliAvailable = false`（唯一的跳过情形）。不中断，继续搭建。
   - **命令存在但提示未登录** → 进入第 2 步自动登录。
   - **命令存在且已登录** → `cliAvailable = true`，跳到第 3 步。
2. **自动登录**（仅「已安装、未登录」时）：运行 `hap auth login`，走**浏览器授权**并等待授权完成。
   - **严禁使用 PAT / token 登录**，必须浏览器发起授权并等待其自动授权成功。
   - 登录成功 → `cliAvailable = true`。
   - 无图形环境等导致浏览器授权无法完成 → 提示用户手动执行 `hap auth login` 完成授权
     （这**不是**「未安装」，不要据此把 `cliAvailable` 记为 false；在用户完成授权前 Step 12 无法回填）。
3. **写入 context**：把 `cliAvailable` 写入 `hap-context.json`（不写 `progress`）。

> **组织一致性与「设为当前应用」不在本步处理**——它们需要 appId（Step 1 之后才有），由 Step 12 完成：
> `cliAvailable = true` 时，Step 12 会先自动把当前组织切到本应用所在组织、并把在建应用设为当前应用，再回填。
> 详见 `build/steps/12_cli_refinement.md`。

---

## 子 agent 可用性预检

在正式进入执行循环之前，**必须先完成以下检查**：

> [!IMPORTANT]
> **本构建流程的执行即代表用户已明确要求使用 sub agent。** 不得因平台默认规则（如"只有用户明确要求时才能使用 sub agent"）而跳过 sub agent 调用。**严禁退回内联执行**。

1. **探测子 agent 功能**：尝试检测当前平台是否支持子 agent。各平台的子 agent 工具名称不同（Antigravity: `invoke_subagent`；Claude Code: `Task`；Codex: `multi_agent_v1.spawn_agent`），子 agent 工具可能不会出现在初始工具列表中，必须先通过工具发现查询，不得仅因初始工具列表未显示就判定子 agent 不可用。
2. **根据检测结果分流**：
   - **子 agent 可用** → 静默通过，继续执行
   - **子 agent 可用但必须用户同意** → 执行下方授权提示流程
   - **子 agent 不可用** → 告知用户将全程使用内联模式执行，搭建后期质量将直线下降，是否继续

#### 授权提示流程

如果检测到当前平台需要用户授权才能使用子 agent，在搭建开始前向用户输出以下说明：

```
ℹ️ 子 agent 授权说明

本次搭建从 Step 4 开始将使用子 agent（子代理）来隔离执行各步骤。
当前平台需要您授权后才能使用此功能。

子 agent 的作用：将复杂步骤委派给独立的子代理执行，避免主对话上下文过载，提高搭建质量。

是否同意使用子 agent？
- 1. 同意：后续步骤将以子 agent 模式执行（推荐）
- 2. 不同意：所有步骤在主对话中执行，但执行后期搭建质量将直线下降
```

等待用户回复后再继续。如果用户不同意，则 Step 4~11 全部退回内联执行。

---

## 执行循环

**按路由表推进**，根据每步标注的执行方式选择内联或 subagent。本流程包含三组并行派发点（详见下方说明）。**所有 `progress` 写入由调度器在验证通过后统一完成，各 step 不写 progress。**

---

### 阶段 1：应用创建（Step 1，内联执行）

Step 1 是轻量的应用和导航分组创建，在主 agent 内执行：

1. 读取 `build/steps/1_create_app.md` → 执行 → 调度器写入 `progress=app_created`

---

### 阶段 1.5：工作表创建与字段刷新（Step 2~3）

Step 2 是整个搭建流程中规则最重的步骤（~400 行规则），**必须使用子 agent 隔离执行**，避免大量 MCP 调用和字段配置数据污染主调度器上下文，确保规则遵守率。

1. 将 Step 2 委派给子 agent → 等待完成 → 调度器写入 `progress=worksheets_created`
2. 读取 `build/steps/3_refresh_fields.md` → 内联执行脚本（一条命令） → 调度器写入 `progress=fields_refreshed`

> **播报**：Step 3 完成后向用户输出：`✅ 基础搭建完成：应用已创建，{N} 张工作表，字段结构已刷新`

---

### 阶段 2：动作、视图与数据（Step 4~6，子 agent 执行）

> [!CAUTION]
> **Step 2、Step 4~11 都必须将任务委派给子 agent。**

#### 并行派发 ①：Step 4 + Step 6（fields_refreshed 后触发）

Step 6（示例数据）仅依赖 `worksheetContext.json` 和 `worksheetIdByName`（Step 3 的产出），与 Step 4/5 无数据依赖。因此：
- Step 3 完成后，同时派发 Step 4 和 Step 6
- Step 4 → Step 5 串行推进，**不等待 Step 6**
- Step 6 作为后台任务运行，在 Step 5 完成前确认完成即可

Step 5 和 Step 6 都完成后，调度器写入 `progress=sample_data_created`。

---

### 阶段 2.5：页面空壳创建（Step 5b，内联执行）

Step 5b 是轻量操作（创建空白页面导航项 + chatbot），为后续三路并行提供 ID 依赖。

读取 `build/steps/5b_create_page_shells.md` → 执行 → 调度器写入 `progress=page_shells_created`

---

### 阶段 3：配置并行（Step 7 + Step 8 + Step 9，子 agent 执行）

#### 并行派发 ②：Step 7 + Step 8 + Step 9（page_shells_created 后触发）

Step 5b 完成后，三者的输入已全部就绪：
- **Step 7**（配置页面组件）：需要 `customPageIdByName`（Step 5b）+ `worksheetContext` + `viewIdByName`
- **Step 8**（创建角色）：需要 `customPageIdByName` + `chatbotIdByName`（Step 5b）+ `worksheetContext` + `viewIdByName`
- **Step 9**（设计工作流）：需要 `worksheetContext` + `viewIdByName` + `roles[]`（来自 hap-plan.json，不依赖 roleContext）

因此：
- Step 5b 完成后，同时派发 Step 7、Step 8、Step 9
- 三者都完成后，调度器写入 `progress=config_completed`

---

### 阶段 4：工作流部署（Step 10 + Step 11，子 agent 执行）

#### 并行派发 ③：Step 10 + Step 11（config_completed 后触发）

Step 10（系统工作流）和 Step 11（自定义动作工作流）均依赖 Step 9 的设计产出，彼此无依赖。因此：
- `config_completed` 后，同时派发 Step 10 和 Step 11
- 两者都完成后，调度器写入 `progress="workflows_deployed"`

---

### 阶段 5：CLI 建后精修（Step 12，内联执行）

Step 12 用 `hap` 命令行工具补 MCP 盖不到的硬缺口。**永不阻断**：CLI 不可用/组织不一致时降级为
「待补清单」，应用仍算搭建成功。

读取 `build/steps/12_cli_refinement.md` → 内联执行 → 调度器写入 `progress="completed"`。

> 本步内联执行（非 subagent）：它依赖构建入口「CLI 自检」写入的 `cliAvailable`，
> 且以对账+少量 CLI 命令为主，上下文开销小。

---

### 子 agent 执行流程

对 Step 4~11 的每一步：

1. **必须将任务委派给子 agent**，使用下方 Prompt 模板
2. **委派成功** → 等待子 agent 完成 → 验证产出数据是否到位（对照全局执行清单）→ 调度器写入 progress
3. **子 agent 不可用** → 退回内联：完整读取步骤文件，在主 agent 内执行

#### 子 agent Prompt 模板

```
你是 HAP 应用搭建执行器，负责执行一个特定的搭建步骤。

## 你的任务
完整阅读步骤文件 `{STEP_FILE_PATH}` 并严格按其要求执行。

## 关键信息
- 应用名称：{appName}
- 项目根目录：{PROJECT_ROOT}
- Skill 目录：{SKILL_DIR}（步骤文件所在的 skill 根目录）
- 方案文件：{PROJECT_ROOT}/apps/{appName}/hap-plan.json
- 进度文件：{PROJECT_ROOT}/apps/{appName}/hap-context.json
- 字段结构：{PROJECT_ROOT}/apps/{appName}/worksheetContext.json（如存在）
- 引用的规则文件：{RULE_FILES}（如果步骤文件引用了共享规则文件，此处填入路径列表；没有则留空）

## 执行要求
1. 先完整读取步骤文件
2. **如果步骤文件引用了其他规则文件（如 `workflow_rules.md`），必须先完整阅读该规则文件后再开始执行**
3. 从 hap-plan.json 读取方案数据
4. 从 hap-context.json 读取已有的 ID 映射
5. 如需字段信息，从 worksheetContext.json 读取（只读）
6. 如步骤文件要求运行脚本（如 generate_fill_templates.py），使用 `{SKILL_DIR}` 定位脚本路径
7. 严格按步骤文件中的规则执行所有操作
8. 所有 MCP 调用必须使用调度器指定的明道云 MCP 服务（服务名称由调度器在委派时传入）
9. 完成后严格按步骤文件中「完成标志」章节的要求写入数据。写入目标可能是 `hap-context.json` 或 `hap-plan.json`，以步骤文件为准。不写 `progress`（由调度器统一管理）
10. 验证步骤文件末尾的 ⛔ 验证断言全部通过
11. 输出**执行问题总结**：列出执行过程中遇到的所有问题（如 API 报错、字段/选项映射失败、节点跳过、重试等），每条包含问题描述和处理方式。如果全程无问题，输出「无异常」
```

---

### 验证与播报（每步通用）

每步完成后：

1. 验证该步骤的交付物字段非空（对照全局执行清单）
2. 验证通过后，**由调度器写入对应的 `progress` 值**到 `hap-context.json`
3. 若验证失败 → 向用户报告错误并停止

**Step 9 特殊验证（调度器必须执行）**：

Step 9 的产出是写入 `hap-plan.json` 而非 `hap-context.json`。调度器必须在子 agent 完成后：

1. 重新读取 `hap-plan.json`，检查每条 `workflows[]` 和 `customActionWorkflows[]` 的 `nodes` 数组是否非空且长度 ≥ 1
2. 检查失败 → **不写入 progress**，向用户报告哪些工作流缺少节点方案，并重新派发 Step 9

**播报节点**（非播报节点静默）：

| 时机 | 输出模板 |
|------|---------| 
| 搭建开始 | `🚀 开始搭建应用【{appName}】…` |
| 阶段 1 完成 | `✅ 基础搭建完成：应用已创建，{N} 张工作表` |
| Step 5b 完成 | `✅ 页面空壳与 AI 助手已创建，开始并行配置…` |
| 阶段 3 完成 | `✅ 页面组件、角色、工作流设计全部完成` |
| 阶段 4 完成 | `✅ 工作流已全部发布，开始建后精修…` |
| Step 12 完成（已回填） | `✅ 建后精修完成：已用 CLI 补齐 {N} 项 MCP 未覆盖的配置` |
| Step 12 完成（降级，hap 未安装） | `ℹ️ 应用已建好；安装并登录 hap-cli 后可补齐 {M} 项增强配置（见摘要）` |
| 全部完成 | 输出完整摘要（见下方「完成」章节） |

---

## 路由表

| progress 值 | 下一步 | 步骤文件 | 执行方式 | 引用规则文件 |
|---|---|---|:---:|---|
| （无/新建） | Step 1：创建应用与导航 | `build/steps/1_create_app.md` | 🔵 内联 | — |
| `app_created` | Step 2：创建工作表 | `build/steps/2_create_worksheets.md` | 🟢 subagent | — |
| `worksheets_created` | Step 3：刷新字段结构 | `build/steps/3_refresh_fields.md` | 🔵 内联 | — |
| `fields_refreshed` | Step 4 + **⚡ Step 6**：并行派发 | `4_create_actions.md` + `6_create_sample_data.md` | 🟢 subagent | — |
| `actions_created` | Step 5：创建视图 | `build/steps/5_create_views.md` | 🟢 subagent | — |
| `sample_data_created` | Step 5b：创建页面空壳与 AI 助手 | `build/steps/5b_create_page_shells.md` | 🔵 内联 | — |
| `page_shells_created` | **⚡ 三路并行** Step 7 + Step 8 + Step 9 | `7_create_pages.md` + `8_create_roles.md` + `9_design_workflows.md` | 🟢 subagent | Step 9 无引用 |
| `config_completed` | **⚡ 并行派发** Step 10 + Step 11 | `10_create_workflows.md` + `11_create_action_workflows.md` | 🟢 subagent | `build/steps/workflow_rules.md` |
| `workflows_deployed` | Step 12：CLI 建后精修对账与回填 | `build/steps/12_cli_refinement.md` | 🔵 内联 | `build/CAPABILITY_MATRIX.md` |

---

## 完成

在标记 `progress="completed"` 之前，**必须回到顶部的「🔒 全局执行清单」逐项核对**。并行派发的步骤须等待全部完成后再推进（详见上方「并行派发策略」）。

确认全部 15 项均已完成后，输出成功摘要：

- 应用名称和链接
- 已创建的工作表数量
- 已写入的示例数据记录条数
- 已创建并发布的工作流数量（系统工作流 + 自定义动作工作流）
- 已创建的角色数量
- 已创建的 AI 助手数量（若有）
- CLI 建后精修结果：已回填 N 项 /（或）待补清单 M 项（CLI 未安装或组织不一致时）

---

## 禁止事项

- **禁止跳过 step 文件**直接凭经验执行
- **禁止伪造 ID**——appId、worksheetId、fieldId、viewId、workflowId 必须来自工具返回值
- **禁止把 MCP 原始返回全文写入 context**——只提取关键 ID 和映射
- **禁止一个 step 修改其他 step 的职责范围**
- **禁止跳步**——必须严格按路由表顺序执行

## 输出格式

- 使用加粗和反引号标注关键名称和数字
- 不使用其他 Markdown 格式
- 错误示例：`❌ Step 2 创建工作表失败：【任务清单】未能创建，原因：xxx`
