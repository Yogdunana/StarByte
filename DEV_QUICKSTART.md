# StarByte 开发者快速上手指引

> **你只需要做两件事：读本文档 + 对 AI 说一句话。**
>
> 仓库地址：https://github.com/Yogdunana/StarByte
> 最后更新：2026-09-04

---

## 一、认领任务（只需一句话）

打开 GitHub Issues 页面，找到一个你想做的 Issue，然后对 AI 说：

```
我要认领 Issue #X
```

就这样。AI 会自动完成以下全部工作：

1. 读取 `TEAM_DEV_GUIDE.md` 了解开发规范
2. 检查该 Issue 的优先级（P0/P1/P2/P3）是否允许你认领
3. 检查你名下是否已有 2 个 Issue（上限）
4. 读取 `REQUIREMENT_CHANGES.md` 检查是否有未处理的需求变更
5. 帮你在 Issue 评论区按规定格式发布认领评论
6. 开始阅读 Issue 详情，进入开发

**你无需全程关注**，仅需在 AI 提问时进行回复即可。

---

## 二、AI 提示词模板（复制粘贴即可）

每次开始新的开发会话时，把以下提示词发给 AI：

```
你是 StarByte 项目的全栈开发 AI 助手。

## 工作准则

1. 每次开发前，必须先读取仓库的 TEAM_DEV_GUIDE.md 文件，严格遵循其中的规范
2. 每次开发前，必须读取 REQUIREMENT_CHANGES.md 检查与你负责的 Issue 相关的未处理变更
3. 认领前检查优先级规则：
   - P0 未全部完成时，禁止认领 P1/P2/P3
   - P1 **核心 8 项**（#3/#4/#5/#17/#18/#19/#20/#21）完成 <50% 时，禁止认领 P2
   - 扩展 P1（#47–#50、#71–#76 等）不计入上述 50% 分母
   - P2 完成 <80% 时，禁止认领 P3
   - 每人最多同时认领 2 个 Issue
   - 禁止同时认领 2 个 P0/P1 高优先级 Issue
4. 认领时，在 Issue 评论区按以下格式发布：
   我来认领 @Yogdunana
   GitHub 账户名：[你的 GitHub 用户名]
   预计完成时间：YYYY-MM-DD
   关联分支：feature/[issue-number]-[brief-description]

## 沟通准则

你不懂的就问我，一问一答。
同时说清楚你为什么要问该问题。
请推荐答案，并说明推荐理由。
直到你对我的目标有明确认知后再开始干活。

## 技术栈

- 后端：Go 1.22 + Gin + GORM + PostgreSQL 16 + Redis 7 + MinIO
- 前端：React 18 + TypeScript 5 + Vite 5 + Redux Toolkit 2 + Ant Design 5
- 鉴权：JWT + Refresh Token
- API：RESTful，前缀 /api/v1/
- 架构：Monorepo，后端四层架构（handler → service → repo → model）

## 项目仓库

https://github.com/Yogdunana/StarByte

现在我要认领 Issue #X（把 X 替换为实际编号）。
```

> **注意**：把 `#X` 替换为你实际要认领的 Issue 编号，把 `[你的 GitHub 用户名]` 替换为你的真实 GitHub 用户名。

---

## 三、认领后的开发流程

AI 认领任务后会自动进入开发状态。流程如下：

```
你说「我要认领 Issue #X」
    ↓
AI 读取 TEAM_DEV_GUIDE.md + Issue 详情
    ↓
AI 检查优先级规则 ← 不通过则告诉你原因，建议换一个
    ↓
AI 检查需求变更 ← 有变更则先处理变更
    ↓
AI 帮你在 Issue 评论区发布认领评论
    ↓
AI 创建分支 feature/X-xxx
    ↓
AI 开始编码（有疑问会问你）
    ↓
AI 提交 PR（按规范填写）
    ↓
你 Review 并合并
```

**关键点**：AI 在编码过程中如果遇到不清楚的地方，会主动问你。你只需要回答问题，AI 会继续工作。

---

## 四、优先级体系速查

| 优先级 | 标签 | 说明 | 认领规则 |
|--------|------|------|----------|
| P0 | `priority:p0` | 核心基础设施 | 必须最先完成，禁止跳过 |
| P1 | `priority:p1` | 核心功能模块 | P0 全部完成后才能认领；门禁只计核心 8 项（#3/#4/#5/#17/#18/#19/#20/#21） |
| P2 | `priority:p2` | 业务功能 | P1 **核心 8 项**完成至少 50% 后可认领 |
| P3 | `priority:p3` | 辅助功能 | P2 完成至少 80% 后可认领 |

**认领限制**：
- 每人最多同时认领 2 个 Issue
- 禁止同时认领 2 个 P0/P1 高优先级 Issue
- 若需放弃认领，在评论区说明并通知 @Yogdunana

---

## 五、需求变更机制

如果项目负责人调整了需求，你的 Issue 会出现以下信号：

1. Issue 标签出现 `req:changed`
2. Issue 评论区有 `[REQ_CHANGE]` 开头的评论
3. `REQUIREMENT_CHANGES.md` 中有你的 Issue 的变更记录

**你的 AI 会自动检查这些信号**。如果发现变更，AI 会：
- 停止当前编码
- 向你展示变更内容
- 询问你是否调整方案
- 适配变更后继续开发

---

## 六、遇到问题时

| 情况 | 做法 |
|------|------|
| AI 建议你换一个 Issue（优先级不够） | 听从建议，选择当前允许认领的 Issue |
| AI 发现需求变更 | 按 AI 提示确认变更，让它适配后继续 |
| 你想放弃已认领的 Issue | 在 Issue 评论区说明，通知 @Yogdunana |
| 编译/运行报错 | 直接把错误信息发给 AI，它会修复 |
| 想了解项目全貌 | 让 AI 读取 `TEAM_DEV_GUIDE.md` 并给你总结 |

---

## 七、快速命令参考

| 你说的话 | AI 会做的事 |
|---------|------------|
| 我要认领 Issue #X | 检查→认领→创建分支→开始开发 |
| 帮我检查 Issue #X 有没有需求变更 | 读取 REQUIREMENT_CHANGES.md + Issue 评论 |
| 帮我提交 PR | 按 PR 规范提交 |
| 帮我看看其他 Issue 的优先级 | 列出可认领的 Issue 并推荐 |
| 我不想做了 | 在评论区发布放弃认领说明 |
