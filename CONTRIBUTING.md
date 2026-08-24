# StarByte 贡献指南

感谢参与 StarByte 项目开发！请仔细阅读本指南，确保开发流程规范。

## 快速上手

### 1. 环境准备

```bash
# 克隆仓库
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte

# 后端环境（需要 Go 1.22+）
cd backend
cp .env.example .env
go mod download

# 前端环境（需要 Node.js 18+）
cd ../frontend
npm install

# 启动基础设施（PostgreSQL、Redis、MinIO）
docker-compose -f ../deploy/docker-compose.yml up -d postgres redis minio

# 运行数据库迁移
cd ../backend && go run cmd/server/main.go migrate

# 启动后端
go run cmd/server/main.go

# 启动前端（新终端）
cd ../frontend && npm run dev
```

### 2. 领取任务

1. 查看 [GitHub Issues](https://github.com/Yogdunana/StarByte/issues) 中的待办任务
2. 选择你感兴趣的 Issue，评论 `我来认领`
3. 等待分配后开始开发

### 3. 开发流程

```bash
# 1. 从 main 拉取最新代码
git checkout main
git pull origin main

# 2. 创建开发分支
git checkout -b feature/your-feature-name

# 3. 开发（建议使用 AI 辅助）
#    开发前必读：docs/dev-guide/ai-prompt-quick-start.md

# 4. 代码检查
#    后端：
cd backend && go fmt ./... && go vet ./...
#    前端：
cd frontend && npm run lint

# 5. 提交代码
git add .
git commit -m "feat: 添加 xxx 功能"

# 6. 推送并创建 PR
git push origin feature/your-feature-name
# 在 GitHub 上创建 Pull Request
```

## 开发规范

### 必读文档

| 文档 | 说明 | 路径 |
|------|------|------|
| 整体架构设计 | 系统架构、技术选型、模块划分 | `docs/specs/00-overall-architecture.md` |
| 工作流引擎设计 | 流程引擎核心设计 | `docs/specs/01-workflow-engine.md` |
| RBAC 权限设计 | 权限系统设计 | `docs/specs/02-rbac-system.md` |
| 后端开发规范 | Go 代码规范、分层架构 | `docs/dev-guide/backend.md` |
| 前端开发规范 | React/TS 代码规范 | `docs/dev-guide/frontend.md` |
| Git 工作流 | 分支管理、提交规范 | `docs/dev-guide/git-workflow.md` |
| PR 提交规范 | PR 模板、审查流程 | `docs/dev-guide/pr-specification.md` |
| AI 开发提示词 | AI 辅助开发必读 | `docs/dev-guide/ai-prompt-quick-start.md` |

### 分支命名

| 类型 | 格式 | 示例 |
|------|------|------|
| 功能开发 | `feature/模块-功能` | `feature/rbac-role-management` |
| Bug 修复 | `fix/模块-问题描述` | `fix/user-login-error` |
| 文档更新 | `docs/内容` | `docs/api-reference` |
| 重构 | `refactor/模块-内容` | `refactor/user-service` |

### 提交信息规范（Conventional Commits）

```
<type>(<scope>): <subject>

<body>

<footer>
```

| type | 说明 | 示例 |
|------|------|------|
| feat | 新功能 | `feat(user): 添加用户注册功能` |
| fix | Bug 修复 | `fix(auth): 修复 Token 刷新失败问题` |
| docs | 文档更新 | `docs: 更新 README` |
| refactor | 代码重构 | `refactor(user): 重构用户服务层` |
| test | 测试相关 | `test(user): 添加用户服务单元测试` |
| chore | 构建/工具 | `chore: 更新依赖版本` |

### 代码质量要求

1. **后端**: 运行 `go fmt ./... && go vet ./...` 无错误
2. **前端**: 运行 `npm run lint` 无错误
3. **测试**: 核心逻辑需有单元测试，覆盖率 ≥ 70%
4. **文档**: 新增 API 需要有 Swagger 注释（handler 文件中）
5. **迁移**: 数据库变更必须通过 migration 文件

### PR 要求

1. PR 标题使用 Conventional Commits 格式
2. PR 描述使用模板（已在 `.github/pull_request_template.md` 中配置）
3. 至少 1 人 Code Review 通过
4. CI 检查全部通过
5. 使用 Squash Merge 合并到 `main`

## AI 辅助开发

### 使用 AI 前必读

在使用 AI（ChatGPT、Claude、GitHub Copilot 等）辅助开发前，**必须**阅读 [AI 开发提示词](docs/dev-guide/ai-prompt-quick-start.md)。

### 核心原则

1. **AI 生成的代码必须人工 review**，不能直接提交
2. AI 可能产生幻觉（生成不存在的 API），必须核实
3. 遵循项目规范是底线，不符合的地方手动修改
4. 安全相关代码（鉴权、权限、支付等）要特别仔细检查
5. 单个文件不超过 300 行，超过需拆分

### GitHub Copilot

项目已配置 `.github/copilot-instructions.md`，Copilot 会自动读取项目规范。

## 项目结构

```
StarByte/
├── backend/                 # Go 后端
│   ├── cmd/server/         # 服务入口
│   ├── internal/            # 业务模块（每个模块独立目录）
│   │   └── user/           # 用户模块（参考实现）
│   ├── pkg/                 # 公共包
│   └── migrations/          # 数据库迁移
├── frontend/                # React 前端
│   └── src/
│       ├── api/             # API 请求
│       ├── components/      # 公共组件
│       ├── pages/           # 页面
│       ├── store/           # Redux 状态
│       └── router/          # 路由
├── deploy/                  # 部署配置
├── docs/                    # 文档
│   ├── specs/              # 设计文档
│   └── dev-guide/          # 开发规范
└── .github/                 # GitHub 配置
    ├── workflows/           # CI/CD
    └── ISSUE_TEMPLATE/     # Issue 模板
```

## 获取帮助

- 查看项目 [Wiki](https://github.com/Yogdunana/StarByte/wiki)（待完善）
- 在 [Discussions](https://github.com/Yogdunana/StarByte/discussions) 中提问
- 提交 [Issue](https://github.com/Yogdunana/StarByte/issues) 报告问题

## License

本项目采用 MIT License。
