# GitHub Copilot Instructions for StarByte

## Project Overview
StarByte is a computer association management system built with a Monorepo architecture.
- Backend: Go 1.22 + Gin + GORM + PostgreSQL 16 + Redis 7 + MinIO
- Frontend: React 18 + TypeScript 5 + Vite 5 + Redux Toolkit 2 + Ant Design 5 + React Flow 11 + ECharts 5

## Architecture Rules

### Backend Module Structure
Each module follows a 4-layer architecture: `handler → service → repo → model`
- `handler/`: HTTP handlers (Gin). Only parameter validation, calling service, returning response.
- `service/`: Business logic. Define interface first, then implement. Transaction control lives here.
- `repo/`: Data access layer. GORM CRUD only, no business logic.
- `model/`: Database models with GORM tags. UUID primary keys.
- `dto/`: Request/Response DTOs. Never return Model directly as response.

### Backend Conventions
- Package names: lowercase, no underscores
- File names: snake_case
- Struct/function names: PascalCase (exported) or camelCase (private)
- Errors: always wrap with `fmt.Errorf("xxx: %w", err)`, never ignore
- Response: use `pkg/response` package (`response.OK`, `response.Error`, `response.Page`)
- Error codes: modular ranges (user 2000-2999, rbac 3000-3999, workflow 4000-4999, etc.)
- Logging: structured zap logging with request_id from context
- Database changes: always via golang-migrate migration files

### Frontend Conventions
- Components: PascalCase file and component names
- Functions/variables: camelCase
- Constants: UPPER_SNAKE_CASE
- TypeScript: prefer interface for objects, type for unions. No `any`, use `unknown` if needed
- Redux: one slice per module, use createAsyncThunk for async, provide select functions
- API: use the axios instance from `src/api/request.ts`, organize by module
- Routes: configured in `src/router/routes.tsx`, support nesting and meta
- Permissions: use `usePermission` hook for UI control, backend must verify
- Styles: CSS Modules for component-level, global CSS for theme

### API Design
- RESTful with `/api/v1/` prefix
- Standard response: `{ code: number, message: string, data: T, request_id: string, timestamp: number }`
- Pagination: `{ list: T[], total: number, page: number, page_size: number }`
- Auth: `Authorization: Bearer <access_token>`

### Git Workflow
- Branch from `main`: `feature/xxx` or `fix/xxx`
- Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- Squash Merge to `main`
- At least 1 Code Review approval before merge

### Key Files to Reference
- `docs/dev-guide/backend.md` - Full backend conventions
- `docs/dev-guide/frontend.md` - Full frontend conventions
- `docs/specs/00-overall-architecture.md` - System architecture
- `docs/specs/01-workflow-engine.md` - Workflow engine design
- `docs/specs/02-rbac-system.md` - RBAC permission design
- `backend/internal/user/` - Reference module implementation
