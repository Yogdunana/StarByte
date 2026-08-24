# StarByte - 计算机协会管理系统

> 面向高校计算机协会的一体化管理平台，支持入会申请、面试流程、人员档案、会议投票、任务流转、实习管理等核心功能。

## 项目简介

StarByte 是专为高校计算机协会设计的综合性管理系统，采用模块化架构，支持 20+ 开发者并行协作开发。系统内置可拖拽流程引擎，灵活适应不断变化的面试和审批流程。

## 技术栈

### 后端
- **语言**: Go 1.22
- **框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **对象存储**: MinIO
- **认证**: JWT + Refresh Token
- **数据库迁移**: golang-migrate

### 前端
- **语言**: TypeScript 5
- **框架**: React 18
- **构建工具**: Vite 5
- **状态管理**: Redux Toolkit 2
- **UI 组件库**: Ant Design 5
- **流程设计**: React Flow 11
- **图表**: ECharts 5
- **路由**: React Router 6

### 基础设施
- **容器化**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **代码规范**: ESLint + Prettier (前端) / gofmt + go vet (后端)

## 项目结构

```
StarByte/
├── backend/                    # 后端服务
│   ├── cmd/
│   │   └── server/            # 服务入口
│   ├── internal/              # 内部业务模块
│   │   ├── user/              # 用户模块
│   │   ├── workflow/          # 流程引擎模块
│   │   ├── member/            # 会员模块
│   │   ├── meeting/           # 会议模块
│   │   ├── task/              # 任务模块
│   │   ├── internship/        # 实习模块
│   │   └── notification/      # 通知模块
│   ├── pkg/                   # 公共包
│   │   ├── config/            # 配置管理
│   │   ├── logger/            # 日志
│   │   ├── response/          # 统一响应
│   │   ├── database/          # 数据库
│   │   ├── redis/             # Redis
│   │   ├── middleware/        # 中间件
│   │   └── events/            # 事件总线
│   ├── migrations/            # 数据库迁移
│   └── Dockerfile
├── frontend/                   # 前端应用
│   ├── src/
│   │   ├── api/               # API 接口
│   │   ├── components/        # 组件
│   │   ├── layouts/           # 布局
│   │   ├── pages/             # 页面
│   │   ├── store/             # Redux 状态
│   │   ├── hooks/             # 自定义 Hooks
│   │   ├── utils/             # 工具函数
│   │   ├── types/             # TypeScript 类型
│   │   ├── router/            # 路由配置
│   │   └── styles/            # 全局样式
│   └── Dockerfile
├── deploy/                     # 部署配置
│   └── docker-compose.yml
├── docs/                       # 项目文档
│   ├── specs/                 # 设计文档
│   └── dev-guide/             # 开发规范
├── .github/                    # GitHub 配置
│   ├── workflows/             # CI/CD
│   ├── ISSUE_TEMPLATE/        # Issue 模板
│   └── pull_request_template.md
└── README.md
```

## 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.22+ (本地开发)
- Node.js 18+ (本地开发)

### 使用 Docker Compose 启动

```bash
# 克隆项目
git clone https://github.com/Yogdunana/StarByte.git
cd StarByte

# 启动所有服务
docker-compose -f deploy/docker-compose.yml up -d

# 查看服务状态
docker-compose -f deploy/docker-compose.yml ps
```

服务启动后访问:
- 前端: http://localhost:3000
- 后端 API: http://localhost:8080/api/v1
- MinIO 控制台: http://localhost:9001 (minioadmin / minioadmin)

### 本地开发

#### 后端开发

```bash
cd backend

# 安装依赖
go mod download

# 配置环境变量（参考 .env.example）
cp .env.example .env

# 启动数据库和 Redis（使用 Docker）
docker-compose -f ../deploy/docker-compose.yml up -d postgres redis minio

# 运行数据库迁移
go run cmd/server/main.go migrate

# 启动服务
go run cmd/server/main.go
```

#### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

访问 http://localhost:5173

## 功能模块

### 一期功能 (Phase 1)

| 模块 | 功能描述 | 状态 |
|------|----------|------|
| 用户认证 | 注册、登录、JWT、权限管理 | 骨架完成 |
| 人员档案 | 用户信息、部门、职位管理 | 规划中 |
| 入会申请 | 会员/干事申请、资料审核 | 规划中 |
| 面试管理 | 面试安排、面试评分、流程配置 | 规划中 |
| 流程引擎 | 可视化流程设计、流程实例管理 | 规划中 |
| 会议管理 | 会议创建、议程管理、签到 | 规划中 |
| 会议投票 | 等权投票、加权投票、匿名投票 | 规划中 |
| 任务流转 | 任务创建、分配、跟踪 | 规划中 |
| 实习管理 | 实习记录、时长统计、排名 | 规划中 |
| 消息通知 | 站内消息、WebSocket 推送 | 规划中 |
| 数据统计 | 会员分布、面试数据、ECharts 图表 | 规划中 |
| 审计日志 | 操作日志、登录日志 | 规划中 |

### 二期功能 (Phase 2)
- 财务管理
- 纪律处分记录
- 合同管理
- 第三方登录（微信扫码等）
- 移动端适配
- 更多通知渠道（邮件、短信、企业微信等）

## 开发规范

> **重要**: 所有开发者在开始编码前，请务必阅读以下文档：

- [后端开发规范](docs/dev-guide/backend.md)
- [前端开发规范](docs/dev-guide/frontend.md)
- [Git 工作流](docs/dev-guide/git-workflow.md)
- [PR 提交规范](docs/dev-guide/pr-specification.md)
- [AI 辅助开发提示词](docs/dev-guide/ai-development-prompt.md)

### 快速回顾

**分支管理**: GitHub Flow
- `main` - 生产环境代码
- `feature/xxx` - 功能开发分支
- `fix/xxx` - Bug 修复分支

**提交 PR**:
1. 从 `main` 拉取新分支
2. 开发完成后提交 PR
3. 至少 1 人 Code Review
4. Squash Merge 到 `main`

**代码质量**:
- 后端: 运行 `go fmt ./... && go vet ./...`
- 前端: 运行 `npm run lint && npm run build`

## 设计文档

- [整体架构设计](docs/specs/00-overall-architecture.md)
- [工作流引擎设计](docs/specs/01-workflow-engine.md)
- [RBAC 权限系统设计](docs/specs/02-rbac-system.md)

## API 文档

启动后端服务后访问: http://localhost:8080/swagger/index.html

> 待接入 Swagger，目前可参考各模块 handler 文件中的注释。

## 贡献指南

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## License

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 联系方式

- 项目地址: [GitHub](https://github.com/Yogdunana/StarByte)
- Issue 反馈: [Issues](https://github.com/Yogdunana/StarByte/issues)
