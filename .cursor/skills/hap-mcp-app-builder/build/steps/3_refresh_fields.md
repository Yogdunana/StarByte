# Step 3：刷新字段结构

你是 HAP 字段结构采集器，负责运行脚本获取所有已创建工作表的完整字段列表并生成 `worksheetContext.json`。

## 输入数据

- `worksheetIdByName`：工作表名称 → worksheetId 映射（来自 `hap-context.json`）
- MCP 配置中的认证 token（`md_pss_id xxx` 格式）

## 执行流程

1. 从当前平台的 MCP 配置文件中提取认证 token（URL 中 `Authorization=` 后的值，`%20` 还原为空格）：
   - Antigravity：`~/.gemini/config/mcp_config.json`
   - Claude Code：`~/.mcp.json`
   - Codex：`~/.codex/config.toml`

2. 运行脚本：
   ```bash
   python3 {SKILL_DIR}/build/scripts/refresh_fields.py \
     --token "md_pss_id xxx" \
     {PROJECT_ROOT}/apps/{appName}/hap-context.json
   ```

3. 脚本会自动：
   - 从 `hap-context.json` 读取 `appId` 和 `worksheetIdByName`
   - 直接调用明道云 REST API 获取每张表的字段结构
   - 标准化字段并写入同目录的 `worksheetContext.json`

4. 更新 `hap-context.json`：不写 `progress`（由调度器统一管理）

**⛔ 验证断言**：`worksheetContext.json` 文件存在且非空，条目数 = 已创建工作表数，每表的 `fields` 数组非空。
