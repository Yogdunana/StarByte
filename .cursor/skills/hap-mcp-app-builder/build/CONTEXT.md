# hap-context.json 结构定义

本文件定义 `{PROJECT_ROOT}/apps/{appName}/hap-context.json` 的完整结构。所有步骤读写 context 时必须遵守此 schema。

## 结构

```json
{
  "appId": "string",
  "appName": "string",
  "org_id": "string",
  "progress": "string (见 PROGRESS.md)",

  "cliAvailable": "boolean (CLI 自检写入：本机 hap 是否已安装且已登录；未登录会在自检阶段自动浏览器登录。组织校正与设当前应用由 Step 12 处理)",

  "sectionIdByName": {
    "分组名称": "sectionId"
  },
  "worksheetIdByName": {
    "工作表名称": "worksheetId"
  },
  "actionIdByName": {
    "动作名称": "actionId"
  },
  "viewIdByName": {
    "工作表名称-视图名称": "viewId"
  },
  "customPageIdByName": {
    "页面名称": "pageId"
  },
  "chatbotIdByName": {
    "助手名称": "chatbotId"
  },
  "roleContext": [
    {
      "roleId": "string",
      "roleName": "string"
    }
  ],
  "customActionWorkflows": [
    {
      "actionName": "string",
      "processId": "string",
      "published": true
    }
  ]
}
```

## 写入规则

- **每个步骤只写入自己负责的字段**，不修改其他步骤的字段
- **只存 ID 映射**，不存 MCP 原始返回的完整对象
- `progress` 字段由每步完成时更新，值域见 `PROGRESS.md`
- 字段结构（fields）存储在独立文件 `worksheetContext.json`，不写入 `hap-context.json`，避免上下文膨胀

## worksheetContext.json 使用契约

`worksheetContext.json` 由 Step 3 写入，**Step 4~10 只读**。

- Step 4~10 中需要的字段 ID、选项 key 等，**必须且只能从 `worksheetContext.json` 查找，严禁重复调用 `get_worksheet_structure`**
- 视图 ID 从 `viewIdByName`（存储在 `hap-context.json`）中获取
- 该文件在 Step 3 写入后**不再修改**
