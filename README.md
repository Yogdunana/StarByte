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
- 前端: http://localhost/ （容器映射 80 端口）
- 后端 API: http://localhost:8080/api/v1
- 健康检查: http://localhost:8080/health 、`/health/ready`
- Metrics: http://localhost:8080/metrics
- MinIO API: http://localhost:9000

### 本地开发

本地只起基础设施，用 **dev** compose（会把 Postgres/Redis/MinIO 端口打到主机）。生产 `docker-compose.yml` 不暴露 5432/6379。

```bash
docker compose -f deploy/docker-compose.dev.yml up -d
```

默认库名 `starbyte_dev`，账号密码 `starbyte` / `starbyte`（见 `backend/configs/config.dev.yaml`）。

#### 后端开发

```bash
cd backend

# 安装依赖
go mod download

# 配置环境变量（参考 .env.example）
cp .env.example .env

# 迁移与种子（需已安装 golang-migrate；库名与 compose.dev 对齐）
# make migrate-up POSTGRES_DB=starbyte_dev
# make seed

# 启动服务（必须编译整个 cmd/server 包，不要只 run main.go）
APP_ENV=dev go run ./cmd/server
```

仓库根目录 `Makefile` 提供统一命令：

```bash
make migrate-up              # 执行全部迁移
make migrate-down            # 回滚最近一次迁移
make migrate-create name=xx  # 新建迁移文件
make seed                    # 幂等写入测试角色/用户/模板
```

默认账号：`admin/admin123`（社长 + 超管）、`test/test123`（会员）。

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

对照 2026-09-06 可用性检查（见 [docs/phase1-readiness.md](docs/phase1-readiness.md)）。账号：`admin/admin123`。

| 模块 | 功能描述 | 状态 |
|------|----------|------|
| 用户认证 | 注册、登录、JWT、Refresh、会话强制下线 | 可用 |
| 人员档案 | 会员档案、部门、数据范围 | 可用（角色/权限管理页仍占位，API 已有） |
| 入会申请 | 会员/干事申请、审核、补充材料 | 可用 |
| 面试管理 | 场次、评分、签到、统计 | 可用 |
| 流程引擎 | 可视化设计器 + 引擎 API | 设计器可用；实例/待办页占位 |
| 会议管理 | 会议、议程、签到 | 可用 |
| 会议投票 | 等权匿名 + 加权 | 可用 |
| 任务流转 | 创建、分配、看板、超时提醒 | 可用 |
| 实习管理 | 记录、时长、排行、配置开关 | 可用 |
| 消息通知 | 站内 + WebSocket + 邮件增强 | 可用 |
| 数据统计 | ECharts 概览、工作台、数据大屏 | 可用 |
| 审计日志 | 操作日志、Trace、归档/报告 | 可用 |
| 文件 / 表单 | MinIO 文件、动态表单引擎 | 可用 |
| 运维探针 | `/health` `/health/ready` `/metrics`、Swagger | 可用 |

财务、纪律处分、合同、OAuth、移动端仍属二期（门禁表 #22/#23/#24 未做产品页）。

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

- [一期可用性检查（2026-09-06）](docs/phase1-readiness.md)
- [整体架构设计](docs/specs/00-overall-architecture.md)
- [工作流引擎设计](docs/specs/01-workflow-engine.md)
- [RBAC 权限系统设计](docs/specs/02-rbac-system.md)

## API 文档

非生产环境启动后端后访问: http://localhost:8080/swagger/index.html

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
