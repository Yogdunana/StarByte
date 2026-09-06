# Step 12：建后精修（CLI 对账与回填）

你是 HAP 应用搭建的**收尾精修执行器**。MCP 已经把应用从零建好，但 MCP 的工具是「创建导向」的，
有少数配置项 MCP 表达不出来（如视图的部分增强配置、角色对 AI 助手的访问权等）。本步用 `hap`
命令行工具（CLI）把这些 **MCP 盖不到的硬缺口**补上。

> [!IMPORTANT]
> 本步**只补硬缺口**——即「MCP 完全做不到」的项。MCP 能做的对象一律不碰，避免同一对象被
> MCP + CLI 双写产生冲突。硬缺口的判定依据是 `build/CAPABILITY_MATRIX.md` 的「CLI-only 硬缺口」列。

> [!CAUTION]
> 本步**永不让搭建失败**。CLI 不可用或与应用不同组织时，跳过回填、输出「待补清单」即可，
> 应用本身已由 MCP 建好，仍算搭建成功。

## 输入数据

- `appId`、`org_id`、各 ID 映射：来自 `hap-context.json`
- `cliAvailable`：来自 `hap-context.json`（构建入口的「CLI 自检」写入）
- `worksheetContext.json`：真实字段/视图结构（只读）
- `hap-plan.json`：方案期望（视图增强配置、角色权限、动作触发条件、节点设计等）
- `build/CAPABILITY_MATRIX.md`：硬缺口清单与各对象的「Step 12 检查点」

## 执行流程

### 步骤 0：确保 CLI 就绪（组织校正 + 设为当前应用）

读 `cliAvailable`：

- **`cliAvailable = false`**（hap 未安装，或已安装但浏览器授权始终未完成）→ 跳到
  「步骤 3：降级输出」。这是**唯一**跳过回填的情形。
- **`cliAvailable = true`** → 做以下「就绪校正」，然后进入步骤 1：

  1. **组织一致性**：运行 `hap auth whoami` 读当前组织，与 `hap-context.json` 的 `org_id` 比对。
     - 不一致 → 运行 `hap auth set-current-org <org_id>` **自动切换**到本应用所在组织。
       （不要让用户手动切换。）
  2. **设为当前应用**：运行 `hap app select <appId>`，把正在搭建的应用设为默认应用。
     - `set-current-org` 会清空默认应用，所以切组织后**必须**重新 `app select`；
       即使组织本就一致也执行此步，使后续回填命令无需反复传 `--app-id`。

### 步骤 1：对账，得出 `cliGaps[]`

> 不依赖任何「各 step 边建边记」的数据——本步在最后**统一对账**重新算出硬缺口。

逐项对照 `build/CAPABILITY_MATRIX.md` 每个对象的「Step 12 检查点」，用 CLI 读出真实结构、与
`hap-plan.json` 期望比对，把「MCP 没做到、且属于 CLI-only 硬缺口」的项收集成运行期列表 `cliGaps[]`。

对账时优先核对硬缺口清单里的高频项（按矩阵排序）：

1. **视图增强配置**：plan 期望的 color / group / filterList / 封面 / quickActions 等，
   逐视图用 `hap --json worksheet view info <ws_id> <view_id>` 读现状，缺的记入 `cliGaps`。
2. **角色 → AI 助手访问权**：plan 中角色应能访问的 AI 助手 vs 角色实际权限（已知历史缺口，必查）。
3. **角色细粒度权限/成员**。
4. **工作流节点深配置**：plan 节点设计意图 vs 实建节点配置。
5. **动作按钮高级配置**（触发条件/可见性）。
6. **特殊字段类型**（`create_worksheet` 不支持、被降级或缺失的字段）。
7. **页面组件细配置**。

每条 `cliGaps[]` 记录：`{ object, id, op, intent, evidence }`
（对象类型、目标 id、要执行的 CLI 操作、方案意图、对账证据=期望 vs 现状）。

> `cliGaps` 为空 → 说明 MCP 已全部覆盖，直接进入「完成」。

### 步骤 2：逐项回填

对每条 `cliGaps[]`，调用对应 `hap` 命令回填。命令写法**直接复用 app 编辑能力的命令字典**
（视图 / 工作流节点 / 角色 / 页面组件 / 字段 / 动作各模块），本步只做桥接，不自创编辑逻辑：

- 各种 id 从 `hap-context.json` 取（appId、worksheetId、viewId、roleId、processId/nodeId…）；
  这些是同后端的服务端 id，CLI 直接可用。
- 改复杂值前**先用读命令导出现状**，在真实结构上改再写回（视图/节点/页面这类整体写回的对象尤其如此）。
- **逐条记录结果**：成功 / 失败（含原因）。单条失败不影响其他项，继续。

> [!CAUTION]
> 回填只针对 `cliGaps[]` 列出的硬缺口。**不要**用 CLI 去重做 MCP 已经建好的对象。

### 步骤 3：降级输出（仅 `cliAvailable = false`，即 hap 未安装/未完成登录）

不回填，改为把按 plan 推断的待补项渲染成**「待补清单」**写入收尾摘要：

> ℹ️ 应用已建好。以下增强配置 MCP 暂未覆盖。安装并登录 hap-cli 后，可用「应用编辑」能力补齐：{清单}
>
> 安装：`pip install hap-cli`；登录：`hap auth login`（浏览器授权）。

清单每项写明：对象、所在工作表/页面、要补什么（自然语言），不要写裸命令。

> 组织不一致**不再**走降级——CLI 可用时已在步骤 0 自动切换组织并设当前应用。

## 完成

不写 `progress`（由调度器统一管理）。调度器在本步通过后写入 `progress="completed"`。

**⛔ 验证断言**：
1. 已读取 `cliAvailable` 并据此分流；
2. `cliAvailable = true` 时：步骤 0 已完成组织校正（必要时 `set-current-org`）+ `app select <appId>`；
   `cliGaps[]` 已对账得出，且每条已尝试回填并记录结果（成功/失败）；
3. `cliAvailable = false` 时：已输出「待补清单」（含安装/登录指引），**未中断搭建**；
4. 本步从未因 CLI 相关问题让整体搭建判为失败。

## 输出要求

- 简洁汇报：本步补了几项、失败几项（附原因）、或降级为待补清单（附清单条数）。
- 失败项与待补项都要让用户能据此手动跟进（后续用「应用编辑」能力自行补）。

## 边界

- 本步**只做 `cliGaps[]` 的回填**，到此为止。
- **开放式精修 / 进一步改动不代劳**：收尾引导用户后续用 `hap` 命令行的「应用编辑」能力
  （从读取应用结构起步）自行修改，本步不主动扩大改动范围。
