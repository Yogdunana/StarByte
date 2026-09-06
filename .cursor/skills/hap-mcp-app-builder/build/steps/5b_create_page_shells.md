# Step 5b：创建自定义页面空壳与 AI 助手

你负责为所有自定义页面创建空白导航项，并创建所有 AI 助手。本步骤只创建对象、获取 ID，**不配置页面组件内容**（组件配置由 Step 7 完成）。

## 输入数据

- `appId`：应用 ID
- `sectionIdByName`：导航分组名称 → sectionId 映射（来自 `hap-context.json`）
- `customPages`：页面规划列表（来自 `hap-plan.json`）
- `aiAssistants`：AI 助手规划列表（来自 `hap-plan.json`，可能为空数组）

## 执行流程

### 阶段 A：创建自定义页面空壳

对每个自定义页面，调用 `create_app_items` 创建空白自定义页面项（挂在指定导航分组下），获得页面 ID：
- `icon`：根据 `pageType` 设定——`dashboard` 传 `"sys_control-panel_traffic"`，`workspace` 传 `"2_3_statistics"`

记录 `customPageIdByName`（格式：`"页面名称" → pageId`）。

**⛔ 验证断言**：`customPageIdByName` 条目数 = plan 中自定义页面数量。

### 阶段 B：创建 AI 助手

1. 读取 `hap-plan.json` 中的 `aiAssistants` 数组
   - 若为空数组或不存在 → 跳过此阶段
2. 对每个 AI 助手，调用 `create_chatbot` 创建：
   - 必填参数：`appId`、`name`、`prompt`、`welcomeMessage`、`presetQuestions`
   - `prompt`：基于助手描述生成简练的系统提示词
   - `presetQuestions`：根据业务场景生成高频预设问题，**必须少于 5 个**
   - `icon`：固定使用 `17_6_reddit`
   - `sectionId`：从 `sectionIdByName` 查找所在导航分组 ID
3. 记录 `chatbotIdByName`（格式：`"助手名称" → chatbotId`）

**⛔ 验证断言**：若 plan 有 AI 助手，则 `chatbotIdByName` 条目数匹配。若 plan 无 AI 助手，直接跳过。

### 完成

更新 `hap-context.json`：写入 `customPageIdByName` 和 `chatbotIdByName`（若有）。不写 `progress`（由调度器统一管理）。
