# 能力归属矩阵（MCP 创建 vs CLI 精修）

> 本表是「何时把活交给 CLI」的唯一事实来源。Step 12（`steps/12_cli_refinement.md`）
> 末尾对账时，依据本表的 **CLI-only 硬缺口** 列识别 MCP 没做到、需用 `hap` CLI 回填的项。
>
> **MCP 列**来自本构建器各 step 实际调用的工具（builder 用到即证明该工具存在，是 MCP 表面的可靠下界）。
> **CLI 列**来自 `hap` 命令表面（细粒度编辑能力，对应 app 编辑相关命令族）。
>
> ⚠️ **校准须知**：标 `🔶待校准` 的格子，需用明道云 MCP server 的真实工具 schema 复核
> （例如 `create_view` 的 payload 到底能不能表达某配置项）。在拿到真实 schema 前，这些按
> 「MCP 可能做不到 → 暂列为硬缺口候选」从严处理，宁可 Step 12 多查一次，也不要漏。

## 图例

- `MCP` —— MCP 工具能做，建构走 MCP，Step 12 不介入
- `CLI-only` —— MCP 做不到的**硬缺口**，进入 Step 12 `cliGaps[]` 由 CLI 回填
- `both` —— 两边都能做；默认走 MCP，不双写
- `🔶待校准` —— 归属取决于 MCP 真实 schema，暂从严当硬缺口候选

---

## 1. 应用与导航

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建应用 | `create_app` | `hap app` 系 | MCP | — |
| 创建导航分组/项 | `create_app_sections` / `create_app_items` | application 模块命令 | MCP | — |
| 修改应用属性/分组（建后） | （无 update 工具） | application 模块命令 | CLI-only | 一般不需要；仅当 plan 要求建后调整分组顺序/图标且 MCP 没表达时 |

## 2. 工作表与字段

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建工作表 | `create_worksheet` | worksheets 模块 | MCP | — |
| 读取字段结构 | `get_worksheet_structure` | `hap worksheet fields --raw` | both | — |
| 建表时的字段集 | `create_worksheet`（随表建字段） | field edit-spec | MCP | — |
| **特殊字段类型**（`create_worksheet` 不支持的类型） | 🔶待校准 | field.add (edit-spec) | 🔶待校准 → CLI-only 候选 | 对账：plan 期望字段类型 vs 实建字段类型，缺失/降级的用 CLI 补 |
| **建后改字段**（改配置/删/重排） | （无字段级 update 工具） | field.update/delete/reorder | CLI-only | 仅当 plan 要求建后调整既有字段时 |

## 3. 自定义动作（按钮）

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建动作 | `batch_create_custom_actions` / `create_custom_actions` | custom-actions 模块 | MCP | — |
| 动作高级配置（enableWhen 等 `create` 未表达项） | 🔶待校准 | custom-action.create/update (action_spec) | 🔶待校准 → CLI-only 候选 | 对账：plan 动作的触发条件/可见性 vs 实建配置 |

## 4. 视图（重点硬缺口区）

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建视图（含 step 5 已列的丰富配置） | `create_view` | views 模块 | MCP | — |
| **建后改视图**（改名/改筛选/改列…） | （无 `update_view` 工具） | `hap worksheet view update` | CLI-only | 仅当 plan 要求建后调视图时 |
| **视图高级配置**（`create_view` payload 表达不出的 editAttrs/advancedSetting 项） | 🔶待校准 | views editAttrs/advancedSetting 全字典 | 🔶待校准 → CLI-only 候选 | 对账：plan 视图期望的增强配置（color/group/filterList/封面等）vs 实建视图，缺的用 CLI 补 |

> 视图是最可能产生硬缺口的对象：MCP 只有 `create_view`，没有 update；凡 `create_view` 一次没配上的，
> 之后只能 CLI 补。校准时重点核对 `create_view` 到底吃哪些配置键。

## 5. 自定义页面与组件

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建页面空壳 | （随 5b，create_custom_page 类） | custom-pages 模块 | MCP | — |
| 配置页面组件 | `update_custom_page` | component edit-spec | both | — |
| **组件细配置**（`update_custom_page` 未表达项） | 🔶待校准 | component.add/update/delete | 🔶待校准 → CLI-only 候选 | 对账：plan 页面组件 vs 实建组件 |

## 6. AI 助手（chatbot）

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建 AI 助手 | `create_chatbot` | （CLI chatbot 相关） | MCP | — |
| 角色对 AI 助手的访问权 | （`create_role` 无 chatbots 字段） | role set-chatbots | CLI-only | 对账：plan 角色应可访问的 AI 助手 vs 实际角色权限（已知历史缺口） |

## 7. 角色与权限

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建角色 | `create_role` | roles 模块 | MCP | — |
| **细粒度权限/成员/AI助手权** | （`create_role` 表达有限） | roles 权限/成员/set-chatbots | CLI-only | 对账：plan 角色权限矩阵 vs 实建角色（尤其 AI 助手访问权，见 §6） |

## 8. 工作流与节点（重点硬缺口区）

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 创建工作流 | `create_process` | workflows 模块 | MCP | — |
| 批量建节点 | `batch_create_process_nodes` | nodes 模块 | MCP | — |
| 读工作流结构 | `get_workflow_structure` | `hap workflow node get` | both | — |
| 删节点/流程 | `delete_process_node` / `delete_process` | nodes/workflows 模块 | both | — |
| 发布 | `publish_process` | `hap workflow publish` | both | — |
| **建后改既有节点配置** | （无节点 update 工具） | nodes 深度字典（8 类节点） | CLI-only | 仅当 plan 要求建后微调既有节点时 |
| **节点深配置**（`batch_create_process_nodes` 表达不出的字段） | 🔶待校准 | nodes.md 深度字典 | 🔶待校准 → CLI-only 候选 | 对账：plan 节点设计意图 vs 实建节点配置 |

## 9. 业务数据（示例数据）

| 对象/操作 | MCP 工具 | CLI 能力 | 归属 | Step 12 检查点 |
|---|---|---|---|---|
| 批量写记录 | `batch_create_records` / `add_record` | `hap worksheet record ...` | MCP | — |
| 读/改/删记录 | `list_records`/`get_record_list`/`update_record`/`delete_record` | record 模块 | both | — |

> 数据不是结合重点：MCP 数据工具齐全，Step 12 不介入示例数据。

---

## CLI-only 硬缺口清单（Step 12 对账主目标）

按"最可能真实存在硬缺口"优先级排序（校准后据实增删）：

1. **视图增强配置**（§4）— `create_view` 表达不出的 color/group/filterList/封面/quickActions 等，建后只能 CLI 补。
2. **角色 → AI 助手访问权**（§6/§7）— `create_role` 无 chatbots 字段，已知历史缺口，必查。
3. **角色细粒度权限/成员**（§7）。
4. **工作流节点深配置**（§8）— `batch_create_process_nodes` 表达不出的节点字段。
5. **动作按钮高级配置**（§3）— 触发条件/可见性。
6. **特殊字段类型**（§2）— `create_worksheet` 不支持的字段类型。
7. **页面组件细配置**（§5）。

## 待校准事项（需明道云 MCP server 真实工具 schema）

- `create_view` 实际接受的配置键全集 → 决定 §4 哪些是真硬缺口。
- `create_worksheet` 支持的字段类型全集 → 决定 §2「特殊字段类型」范围。
- `batch_create_process_nodes` 的节点配置字段全集 → 决定 §8 节点深配置缺口。
- `create_role` / `update_custom_page` / `create_custom_actions` 的 payload 全集 → 决定 §3/§5/§7 缺口。

> 校准方式：拿到 MCP server 工具 schema 后，逐项把 `🔶待校准` 改判为 `MCP`（能表达）或 `CLI-only`（确为硬缺口），并据此精简上面的硬缺口清单。
