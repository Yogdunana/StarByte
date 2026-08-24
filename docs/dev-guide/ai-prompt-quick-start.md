# StarByte AI 开发提示词（必读）

> **所有开发者在使用 AI 辅助开发前，必须先阅读本文档，并将对应提示词粘贴到 AI 对话中。**

## 快速开始

### 第一步：选择你的开发场景提示词

| 场景 | 提示词文件 | 说明 |
|------|-----------|------|
| 后端开发 | [backend-prompt.md](./prompts/backend-prompt.md) | Go + Gin + GORM 后端开发 |
| 前端开发 | [frontend-prompt.md](./prompts/frontend-prompt.md) | React + TypeScript 前端开发 |
| 全栈开发 | [fullstack-prompt.md](./prompts/fullstack-prompt.md) | 前后端联合开发 |
| Code Review | [review-prompt.md](./prompts/review-prompt.md) | 代码审查 |
| 流程引擎 | [workflow-prompt.md](./prompts/workflow-prompt.md) | 流程引擎专项开发 |

### 第二步：复制提示词到 AI 对话

将提示词内容粘贴到 AI 对话的**最前面**（系统提示词位置），然后再描述你的具体需求。

### 第三步：开发流程

```
1. 在 GitHub Issues 中找到你要开发的任务
2. 评论 "我来认领" 领取任务
3. 从 main 拉取新分支：git checkout -b feature/xxx
4. 粘贴 AI 提示词 + 描述需求 → 让 AI 生成代码
5. 人工 review AI 生成的代码
6. 本地测试通过后提交
7. 提交 PR（使用 PR 模板）
8. 至少 1 人 Code Review 通过后合并
```

---

## 通用 AI 提示词（精简版，可直接复制）

```
你是 StarByte 项目的高级开发工程师。StarByte 是一个计算机协会管理系统，采用 Monorepo 架构。

## 技术栈
- 后端：Go 1.22 + Gin + GORM + PostgreSQL 16 + Redis 7 + MinIO
- 前端：React 18 + TypeScript 5 + Vite 5 + Redux Toolkit 2 + Ant Design 5 + React Flow 11 + ECharts 5
- 鉴权：JWT + Refresh Token
- API：RESTful，前缀 /api/v1/
- 部署：Docker Compose

## 仓库地址
https://github.com/Yogdunana/StarByte

## 关键规范
1. 后端模块四层架构：handler → service → repo → model，禁止反向依赖
2. 前端组件用 PascalCase，函数/变量用 camelCase，常量用 UPPER_SNAKE_CASE
3. 主键统一使用 UUID
4. 统一响应格式：{ code, message, data, request_id, timestamp }
5. 错误码按模块分段（用户 2000-2999，权限 3000-3999，流程 4000-4999...）
6. 所有写操作必须有审计日志
7. 权限校验在后端做，前端只做 UI 控制
8. 数据库变更必须通过 migration 文件
9. 提交前运行：后端 go fmt + go vet，前端 npm run lint
10. PR 使用 Squash Merge，提交信息用 Conventional Commits

## 开发前必读文档
- docs/dev-guide/backend.md - 后端开发规范
- docs/dev-guide/frontend.md - 前端开发规范
- docs/dev-guide/git-workflow.md - Git 工作流
- docs/dev-guide/pr-specification.md - PR 规范
- docs/specs/ - 架构设计文档

请先理解以上规范，然后根据我的需求进行开发。生成代码后请说明设计思路和注意事项。
```

---

## 使用示例

### 示例 1：开发新功能

```
[粘贴上方通用提示词]

我要开发「会议投票」功能，需求：
1. 支持等权投票（一人一票，支持匿名）
2. 支持加权投票（按职务配置权重）
3. 实时显示投票结果

请先给出 API 设计和数据库表设计，确认后再写代码。
```

### 示例 2：修复 Bug

```
[粘贴上方通用提示词]

用户登录接口有 bug：当密码包含特殊字符时登录失败。
报错信息：Error 1064: You have an error in your SQL syntax
相关代码：backend/internal/user/service/user_service.go

请分析原因并给出修复方案。
```

### 示例 3：Code Review

```
[粘贴 review-prompt.md 中的提示词]

请审查以下代码：
[粘贴代码]
```

---

## 注意事项

1. **AI 生成的代码必须人工 review**，不要直接提交
2. AI 可能产生幻觉（生成不存在的 API），要核实
3. 安全相关代码（鉴权、权限）要特别仔细检查
4. 遵循项目规范是底线，不符合的地方要手动修改
5. 遇到不确定的，查官方文档或问团队
6. 不要让 AI 生成超过 300 行的单个文件，要拆分
