# 前端开发规范

> **版本**: v1.0
> **日期**: 2026-08-24
> **适用范围**: StarByte 前端 React 项目

---

## 1. 技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| React | 18.x | UI 框架 |
| TypeScript | 5.x | 类型安全 |
| Vite | 5.x | 构建工具 |
| Redux Toolkit | 2.x | 状态管理 |
| React Router | 6.x | 路由 |
| Ant Design | 5.x | UI 组件库 |
| React Flow | 11.x | 流程设计器 |
| ECharts | 5.x | 图表库 |
| Axios | 1.x | HTTP 客户端 |
| ESLint | - | 代码检查 |
| Prettier | - | 代码格式化 |

---

## 2. 项目结构规范

### 2.1 目录结构

```
src/
├── api/                # API 请求封装（按模块组织）
│   ├── request.ts      # Axios 实例封装
│   ├── user.ts
│   ├── member.ts
│   └── ...
├── assets/             # 静态资源
│   ├── images/
│   ├── icons/
│   └── styles/
├── components/         # 公共组件
│   ├── common/         # 通用组件（如 PageContainer、StatusTag）
│   ├── business/       # 业务组件（跨页面复用）
│   └── layout/         # 布局相关组件
├── layouts/            # 布局组件
│   ├── MainLayout.tsx
│   └── BlankLayout.tsx
├── pages/              # 页面（按业务模块组织）
│   ├── login/
│   ├── dashboard/
│   ├── user/
│   │   ├── components/     # 页面私有组件
│   │   ├── UserList.tsx
│   │   └── UserDetail.tsx
│   ├── member/
│   ├── interview/
│   ├── meeting/
│   ├── task/
│   ├── workflow/
│   ├── internship/
│   ├── stats/
│   └── system/
├── store/              # Redux Store
│   ├── index.ts
│   └── slices/         # 按模块拆分 slice
│       ├── userSlice.ts
│       ├── authSlice.ts
│       └── ...
├── hooks/              # 自定义 Hooks
│   ├── usePermission.ts
│   ├── useRequest.ts
│   └── ...
├── utils/              # 工具函数
│   ├── request.ts      # 请求工具
│   ├── storage.ts      # 本地存储
│   ├── date.ts         # 日期处理
│   └── ...
├── types/              # TypeScript 类型定义
│   ├── api.ts          # API 相关类型
│   ├── user.ts         # 用户相关类型
│   └── ...
├── router/             # 路由配置
│   ├── index.tsx
│   └── routes.tsx
├── styles/             # 全局样式
│   ├── global.css
│   └── variables.css
├── App.tsx
└── main.tsx
```

### 2.2 组件文件命名
- 组件使用大驼峰（PascalCase）：`UserList.tsx`
- 普通工具函数使用小驼峰（camelCase）：`useRequest.ts`
- 样式文件：`ComponentName.module.css`（CSS Modules）或与组件同名

---

## 3. TypeScript 规范

### 3.1 类型定义
- 优先使用 `interface` 定义对象类型
- 使用 `type` 定义联合类型、交叉类型、工具类型
- 类型定义放在 `types/` 目录或组件文件顶部
- 禁止使用 `any`，除非万不得已（用 `unknown` 代替）

```ts
// 接口定义
interface User {
  id: string;
  username: string;
  realName: string;
  status: number;
  departmentId?: string;
}

// 联合类型
type UserStatus = 0 | 1 | 2;

// API 响应类型
interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  requestId: string;
  timestamp: number;
}

// 分页请求
interface PageRequest {
  page: number;
  pageSize: number;
  keyword?: string;
}

// 分页响应
interface PageResponse<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}
```

### 3.2 类型断言
- 尽量避免类型断言
- 优先使用类型守卫（Type Guard）
- 断言时使用 `as` 语法

---

## 4. React 组件规范

### 4.1 组件写法
- 使用函数组件 + Hooks
- 使用 TypeScript 定义 Props 类型
- 组件名称和文件名保持一致（大驼峰）

```tsx
import React from 'react';
import { Button } from 'antd';

interface UserCardProps {
  user: User;
  onEdit?: (user: User) => void;
  onDelete?: (id: string) => void;
}

const UserCard: React.FC<UserCardProps> = ({ user, onEdit, onDelete }) => {
  return (
    <div className="user-card">
      <h3>{user.realName}</h3>
      <p>{user.username}</p>
      <div>
        <Button onClick={() => onEdit?.(user)}>编辑</Button>
        <Button danger onClick={() => onDelete?.(user.id)}>删除</Button>
      </div>
    </div>
  );
};

export default UserCard;
```

### 4.2 Hooks 使用规范
- 只在函数组件顶层调用 Hooks
- 只在 React 函数或自定义 Hook 中调用 Hooks
- 自定义 Hook 以 `use` 开头
- useEffect 的依赖数组要完整

```tsx
// 自定义 Hook
function useUserList(params: ListUserParams) {
  const [list, setList] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);

  const fetchList = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getUserList(params);
      setList(res.data.list);
      setTotal(res.data.total);
    } finally {
      setLoading(false);
    }
  }, [params]);

  useEffect(() => {
    fetchList();
  }, [fetchList]);

  return { list, loading, total, refresh: fetchList };
}
```

### 4.3 组件拆分原则
- 单个组件不超过 300 行
- 超过则拆分为子组件
- 复用逻辑抽成自定义 Hook
- 按职责拆分：展示组件、容器组件

---

## 5. 状态管理规范（Redux Toolkit）

### 5.1 Slice 组织
- 按业务模块拆分 slice
- 每个 slice 放在 `store/slices/` 目录下
- slice 名称与模块名一致

```typescript
// store/slices/userSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { getUserList, User } from '@/api/user';
import { RootState } from '..';

interface UserState {
  list: User[];
  total: number;
  loading: boolean;
  currentUser: User | null;
}

const initialState: UserState = {
  list: [],
  total: 0,
  loading: false,
  currentUser: null,
};

// 异步 thunk
export const fetchUserList = createAsyncThunk(
  'user/fetchList',
  async (params: ListUserParams) => {
    const res = await getUserList(params);
    return res.data;
  }
);

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setCurrentUser(state, action: PayloadAction<User | null>) {
      state.currentUser = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchUserList.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchUserList.fulfilled, (state, action) => {
        state.loading = false;
        state.list = action.payload.list;
        state.total = action.payload.total;
      })
      .addCase(fetchUserList.rejected, (state) => {
        state.loading = false;
      });
  },
});

export const { setCurrentUser } = userSlice.actions;

export const selectUserList = (state: RootState) => state.user.list;
export const selectUserLoading = (state: RootState) => state.user.loading;

export default userSlice.reducer;
```

### 5.2 Store 配置
```typescript
// store/index.ts
import { configureStore } from '@reduxjs/toolkit';
import userReducer from './slices/userSlice';
import authReducer from './slices/authSlice';

export const store = configureStore({
  reducer: {
    user: userReducer,
    auth: authReducer,
    // ... 其他 slice
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

### 5.3 使用原则
- 全局共享状态放 Redux（用户信息、权限、主题等）
- 组件内部状态用 useState/useReducer
- 表单状态用 Form 组件管理（Ant Design Form）
- 页面级状态如果不需要共享，用 useState 就够了，不要全塞 Redux

---

## 6. 路由规范

### 6.1 路由配置
- 使用 React Router v6
- 路由配置集中管理，放在 `router/routes.tsx`
- 路由按模块组织，支持嵌套路由

```tsx
// router/routes.tsx
import { Navigate } from 'react-router-dom';
import MainLayout from '@/layouts/MainLayout';
import Login from '@/pages/login/Login';
import Dashboard from '@/pages/dashboard/Dashboard';
import UserList from '@/pages/user/UserList';
import MemberList from '@/pages/member/MemberList';
// ...

export const routes = [
  {
    path: '/login',
    element: <Login />,
    meta: { title: '登录', public: true },
  },
  {
    path: '/',
    element: <MainLayout />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      {
        path: 'dashboard',
        element: <Dashboard />,
        meta: { title: '工作台', icon: 'dashboard' },
      },
      {
        path: 'user',
        children: [
          {
            index: true,
            element: <UserList />,
            meta: { title: '用户管理', permission: 'user:read' },
          },
        ],
      },
      // ... 其他路由
    ],
  },
];
```

### 6.2 路由权限
- 通过路由守卫组件实现权限控制
- 无权限时跳转到 403 页面
- 菜单根据权限动态生成

---

## 7. API 请求规范

### 7.1 请求封装
- 统一使用 Axios 封装
- 请求拦截器：添加 Token、请求 ID
- 响应拦截器：统一处理错误、Token 刷新

```typescript
// api/request.ts
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { message } from 'antd';
import { getToken, getRefreshToken, setToken } from '@/utils/storage';

const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    // 添加请求ID
    config.headers['X-Request-ID'] = generateRequestId();
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { code, message: msg, data } = response.data;
    if (code === 0) {
      return data;
    }
    // 业务错误
    message.error(msg || '请求失败');
    return Promise.reject(new Error(msg || '请求失败'));
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token 过期，尝试刷新或跳转登录
      handleTokenExpired();
    } else {
      message.error(error.message || '网络错误');
    }
    return Promise.reject(error);
  }
);

export default request;
```

### 7.2 API 组织
- 按模块组织 API 函数
- 每个模块一个文件
- 统一导出类型定义

```typescript
// api/user.ts
import request from './request';
import type { User, ListUserParams, CreateUserParams, UpdateUserParams } from '@/types/user';

// 获取用户列表
export function getUserList(params: ListUserParams) {
  return request.get<PageResponse<User>>('/users', { params });
}

// 获取用户详情
export function getUserDetail(id: string) {
  return request.get<User>(`/users/${id}`);
}

// 创建用户
export function createUser(data: CreateUserParams) {
  return request.post<User>('/users', data);
}

// 更新用户
export function updateUser(id: string, data: UpdateUserParams) {
  return request.put<User>(`/users/${id}`, data);
}

// 删除用户
export function deleteUser(id: string) {
  return request.delete(`/users/${id}`);
}
```

---

## 8. 样式规范

### 8.1 方案选择
- 优先使用 Ant Design 的样式和主题定制
- 组件级样式使用 CSS Modules
- 全局样式放在 `styles/global.css`
- 使用 CSS 变量定义主题色

### 8.2 命名规范
- CSS Modules 使用小驼峰：`.userCard { }`
- 全局样式使用 kebab-case：`.user-card { }`
- BEM 可选，不强制

```tsx
// UserCard.module.css
.container {
  padding: 16px;
  border-radius: 8px;
  background: #fff;
}

.title {
  font-size: 16px;
  font-weight: 600;
}

// UserCard.tsx
import styles from './UserCard.module.css';

const UserCard: React.FC = () => (
  <div className={styles.container}>
    <h3 className={styles.title}>标题</h3>
  </div>
);
```

---

## 9. 权限控制规范

### 9.1 权限 Hook
```tsx
// hooks/usePermission.ts
import { useMemo } from 'react';
import { useSelector } from 'react-redux';
import { selectPermissions } from '@/store/slices/authSlice';

export function usePermission(permissionCode: string): boolean {
  const permissions = useSelector(selectPermissions);
  return useMemo(() => {
    return permissions.includes(permissionCode);
  }, [permissions, permissionCode]);
}

// 批量检查权限
export function usePermissions(permissionCodes: string[]): boolean[] {
  const permissions = useSelector(selectPermissions);
  return useMemo(() => {
    return permissionCodes.map(code => permissions.includes(code));
  }, [permissions, permissionCodes]);
}
```

### 9.2 权限组件
```tsx
// components/common/PermissionButton.tsx
import React from 'react';
import { Button, ButtonProps } from 'antd';
import { usePermission } from '@/hooks/usePermission';

interface PermissionButtonProps extends ButtonProps {
  permission: string;
  children: React.ReactNode;
}

const PermissionButton: React.FC<PermissionButtonProps> = ({
  permission,
  children,
  ...rest
}) => {
  const hasPermission = usePermission(permission);
  if (!hasPermission) return null;
  return <Button {...rest}>{children}</Button>;
};

export default PermissionButton;
```

---

## 10. 代码检查与格式化

### 10.1 ESLint
- 使用 ESLint + TypeScript ESLint
- 配置文件：`.eslintrc.cjs`

### 10.2 Prettier
- 使用 Prettier 统一代码格式
- 配置文件：`.prettierrc`

### 10.3 提交前检查
- 使用 husky + lint-staged
- 提交前自动运行 ESLint 和 Prettier

---

## 11. 命名规范总结

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserList.tsx` |
| 工具/Hook 文件 | camelCase | `useRequest.ts` |
| 组件名 | PascalCase | `UserList` |
| 函数/变量 | camelCase | `handleClick` |
| 常量 | UPPER_SNAKE_CASE | `MAX_PAGE_SIZE` |
| 类型/接口 | PascalCase | `User`, `UserService` |
| 样式类（CSS Modules） | camelCase | `.userCard` |
| 全局样式类 | kebab-case | `.user-card` |
| Slice 名称 | camelCase | `userSlice` |
| Action 类型 | domain/action | `user/fetchList` |

---

## 12. 性能优化建议

1. **列表渲染**：长列表使用虚拟滚动（如 `react-window`）
2. **Memo 优化**：合理使用 `React.memo`、`useMemo`、`useCallback`
3. **代码分割**：路由级懒加载，`React.lazy` + `Suspense`
4. **图片优化**：使用 WebP 格式、懒加载、CDN
5. **Bundle 分析**：定期使用 `rollup-plugin-visualizer` 分析包体积
6. **请求缓存**：合理使用 React Query/SWR（如引入）或缓存层
