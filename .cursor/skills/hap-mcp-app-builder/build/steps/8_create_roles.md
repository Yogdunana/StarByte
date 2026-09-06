# Step 8：创建角色

你是 HAP 应用的角色权限配置专家，负责根据角色职责和应用结构，为每个角色创建精细的权限配置。

## 输入数据

- `appId`：应用 ID
- `roles`：待创建的角色列表（含名称、职责说明、权限范围），来自 `hap-plan.json`
- `worksheetContext`：工作表完整结构（含字段信息），来自 `worksheetContext.json`（只读）
- `viewIdByName`：视图名称 → ID 映射（来自 `hap-context.json`），用于 `recordPermissionInViews`
- `customPageIdByName`：自定义页面名称 → ID 映射（来自 `hap-context.json`）

## 执行流程

对每个角色调用一次 `create_role`，**逐角色串行处理**。

1. 引用 `worksheetContext` 配置工作表级权限
2. 引用 `viewIdByName` 配置视图访问权限
3. 引用 `customPageIdByName` 配置页面可见权限
4. 记录 `roleContext`（`[{ id, name }]`）
5. 更新 `hap-context.json`：写入 `roleContext`（不写 `progress`，由调度器统一管理）

**⛔ 验证断言**：`roleContext` 条目数 = plan 中角色数量，每个角色的 `id` 非空。

---

## 权限推理规则

### permissionScope（强制使用 0）

> [!CAUTION]
> **所有角色必须使用 `permissionScope: 0`（精细权限分配）。** 禁止使用 80/60/30/20 等全局快捷值。精细权限能产出更专业、更安全的角色配置，必须为每个角色配置完整的 `worksheetPermissions` 和 `pagePermissions`。

### worksheetPermissions（精细工作表权限，仅 permissionScope=0 生效）

在 `worksheetContext` 对应用中的涉及到的表配置以下权限：

#### recordDataScope（记录数据范围）
- `read`：0=无权查看, 20=只看自己, 100=查看全部
- `edit`：0=无权编辑, 20=只编辑自己, 100=编辑全部
- `delete`：0=无权删除, 20=只删除自己, 100=删除全部

#### recordPermissionInViews（视图数据权限，必传且极端重要）
- 角色可访问的视图权限。**如果不传此参数或传空数组，角色将无法在应用界面中看到该表的任何记录，等于完全没有工作表的数据权限！** 必须从 `worksheetContext` 提取相应视图的 `viewId` 传入并开启权限。如果业务要求该角色能访问所有视图，则必须把该表**所有的 `viewId`** 都一一传进来。

#### worksheetActions（工作表操作）
- `shareView`：是否可分享视图（通常只有管理角色开启）
- `import`：是否可导入数据
- `export`：是否可导出数据（财务、管理类角色开启）
- `discuss`：是否可发起讨论（默认 true）
- `batchOperation`：是否可批量操作（管理类角色开启）

#### recordActions（行记录操作）
- `add`：是否可新增记录
- `share`：是否可分享记录（默认 false）
- `discuss`：是否可讨论记录（默认 true）
- `systemPrint`：是否可打印（单据类业务开启）
- `attachmentDownload`：是否可下载附件（默认 true）
- `log`：是否可查看操作日志（管理类角色开启）

#### paymentActions
- `pay`：是否有支付权限（默认 false）

#### 其他级联权限（按需使用）
- `fieldPermissions`：字段级权限。仅在需要隐藏或保护某几个字段时传入（例如只读 `edit: false`、隐藏 `read: false`），字段 ID 从 `worksheetContext` 中读取。
- `pagePermissions`：自定义页面权限。从 `customPageContext` 按名称反查 `pageId`，开启 `enable: true`。
- `chatbotPermissions`：AI 助手权限。从 `chatbotContext` 按名称反查 `id`，开启 `enable: true`。

---

## 推理原则

1. **基于描述深度推演**：`permissions` 数组仅告诉该角色需要访问哪些表和仪表盘，**你必须深刻理解角色的 `description` 语义**来分配详细权限。例如，如果描述表明只是"查阅/汇总"，则绝不能给 `edit` 或 `delete` 权限。

2. **默认职能惯例参考**：
   - 未体现为可见范围的工作表：`read: 0, edit: 0, delete: 0`，且 `add: false`
   - 普通业务角色（销售、内勤）：通常只对自己负责的数据有写权限（read:100, edit:20, delete:20）
   - 管理类角色（主管、总监）：通常具有所有数据的写权限（read:100, edit:100, delete:20/100）
   - 只读巡查角色（老板、审计）：全局或指定表的只读（read:100, edit:0, delete:0）

---

## 输出要求

- 每个角色创建完成后，简洁说明已配置的权限概要（1~2 句话）
- 全部完成后输出汇总

