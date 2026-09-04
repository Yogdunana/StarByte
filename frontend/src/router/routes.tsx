import React, { lazy, Suspense } from 'react';
import { Navigate } from 'react-router-dom';
import type { RouteObject } from 'react-router-dom';

// 布局组件
import MainLayout from '@/layouts/MainLayout/MainLayout';

// 路由守卫
import AuthRoute from '@/router/guards/AuthRoute';
import PermissionRoute from '@/router/guards/PermissionRoute';

// 页面组件
const Login = lazy(() => import('@/pages/login/Login'));
const Dashboard = lazy(() => import('@/pages/dashboard/Dashboard'));
const UserList = lazy(() => import('@/pages/user/UserList'));
const NotificationList = lazy(() => import('@/pages/notification/NotificationList'));
const TemplateList = lazy(() => import('@/pages/notification/TemplateList'));
const AuditList = lazy(() => import('@/pages/system/audit/AuditList'));
const WorkflowDesigner = lazy(() => import('@/pages/workflow/designer/DesignerPage'));
const Forbidden = lazy(() => import('@/pages/error/Forbidden'));

// 懒加载 fallback
const LoadingFallback: React.FC = () => (
  <div style={{ padding: 24, textAlign: 'center' }}>加载中...</div>
);

// 懒加载包装器（无权限守卫）
const lazyWrap = (Component: React.LazyExoticComponent<React.FC>) => (
  <Suspense fallback={<LoadingFallback />}>
    <Component />
  </Suspense>
);

// 带权限守卫的懒加载包装器
const lazyGuarded = (
  Component: React.LazyExoticComponent<React.FC>,
  permission?: string,
) => {
  const element = lazyWrap(Component);
  return permission ? (
    <PermissionRoute permission={permission}>{element}</PermissionRoute>
  ) : element;
};

// 带权限守卫的内联元素包装器
const guarded = (element: React.ReactNode, permission?: string) =>
  permission ? (
    <PermissionRoute permission={permission}>{element}</PermissionRoute>
  ) : element;

// 路由元信息类型
export interface RouteMeta {
  title?: string;
  icon?: string;
  permission?: string; // 权限码
  public?: boolean; // 是否公开页面（不需要登录）
  hidden?: boolean; // 是否在菜单中隐藏
}

// 扩展 RouteObject 类型
export interface AppRouteObject extends Omit<RouteObject, 'children'> {
  meta?: RouteMeta;
  children?: AppRouteObject[];
}

const routes: AppRouteObject[] = [
  {
    path: '/login',
    element: lazyWrap(Login),
    meta: { title: '登录', public: true, hidden: true },
  },
  {
    path: '/',
    element: <AuthRoute><MainLayout /></AuthRoute>,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      {
        path: '403',
        element: lazyWrap(Forbidden),
        meta: { title: '无权限', hidden: true },
      },
      {
        path: 'dashboard',
        element: lazyGuarded(Dashboard),
        meta: { title: '工作台', icon: 'DashboardOutlined' },
      },
      {
        path: 'user',
        meta: { title: '用户管理', icon: 'UserOutlined' },
        children: [
          {
            index: true,
            path: 'list',
            element: lazyGuarded(UserList, 'user:read'),
            meta: { title: '用户列表', permission: 'user:read' },
          },
        ],
      },
      {
        path: 'notification',
        meta: { title: '通知管理', icon: 'BellOutlined' },
        children: [
          {
            path: 'list',
            element: lazyWrap(NotificationList),
            meta: { title: '通知列表' },
          },
          {
            path: 'templates',
            element: lazyGuarded(TemplateList, 'notification:template:read'),
            meta: { title: '模板管理', permission: 'notification:template:read' },
          },
        ],
      },
      {
        path: 'member',
        meta: { title: '会员管理', icon: 'TeamOutlined' },
        children: [
          {
            path: 'application',
            element: <div style={{ padding: 24 }}>入会申请（开发中）</div>,
            meta: { title: '入会申请' },
          },
          {
            path: 'list',
            element: <div style={{ padding: 24 }}>会员档案（开发中）</div>,
            meta: { title: '会员档案' },
          },
        ],
      },
      {
        path: 'interview',
        meta: { title: '面试管理', icon: 'ScheduleOutlined' },
        children: [
          {
            path: 'list',
            element: <div style={{ padding: 24 }}>面试安排（开发中）</div>,
            meta: { title: '面试安排' },
          },
          {
            path: 'flow',
            element: <div style={{ padding: 24 }}>面试流程（开发中）</div>,
            meta: { title: '面试流程' },
          },
        ],
      },
      {
        path: 'meeting',
        meta: { title: '会议管理', icon: 'CalendarOutlined' },
        children: [
          {
            path: 'list',
            element: <div style={{ padding: 24 }}>会议列表（开发中）</div>,
            meta: { title: '会议列表' },
          },
          {
            path: 'vote',
            element: <div style={{ padding: 24 }}>投票管理（开发中）</div>,
            meta: { title: '投票管理' },
          },
        ],
      },
      {
        path: 'task',
        meta: { title: '任务流转', icon: 'CheckCircleOutlined' },
        children: [
          {
            path: 'list',
            element: <div style={{ padding: 24 }}>任务列表（开发中）</div>,
            meta: { title: '任务列表' },
          },
          {
            path: 'my',
            element: <div style={{ padding: 24 }}>我的任务（开发中）</div>,
            meta: { title: '我的任务' },
          },
        ],
      },
      {
        path: 'workflow',
        meta: { title: '流程管理', icon: 'ApartmentOutlined' },
        children: [
          {
            path: 'designer',
            element: lazyWrap(WorkflowDesigner),
            meta: { title: '流程设计' },
          },
          {
            path: 'designer/:id',
            element: lazyWrap(WorkflowDesigner),
            meta: { title: '流程设计', hidden: true },
          },
          {
            path: 'instances',
            element: <div style={{ padding: 24 }}>流程实例（开发中）</div>,
            meta: { title: '流程实例' },
          },
          {
            path: 'todo',
            element: <div style={{ padding: 24 }}>我的待办（开发中）</div>,
            meta: { title: '我的待办' },
          },
        ],
      },
      {
        path: 'internship',
        meta: { title: '实习管理', icon: 'ReadOutlined' },
        children: [
          {
            path: 'list',
            element: <div style={{ padding: 24 }}>实习记录（开发中）</div>,
            meta: { title: '实习记录' },
          },
          {
            path: 'stats',
            element: <div style={{ padding: 24 }}>实习统计（开发中）</div>,
            meta: { title: '实习统计' },
          },
        ],
      },
      {
        path: 'stats',
        meta: { title: '数据统计', icon: 'BarChartOutlined' },
        children: [
          {
            path: 'overview',
            element: <div style={{ padding: 24 }}>统计概览（开发中）</div>,
            meta: { title: '统计概览' },
          },
        ],
      },
      {
        path: 'system',
        meta: { title: '系统管理', icon: 'SettingOutlined' },
        children: [
          {
            path: 'role',
            element: guarded(<div style={{ padding: 24 }}>角色管理（开发中）</div>, 'role:read'),
            meta: { title: '角色管理', permission: 'role:read' },
          },
          {
            path: 'permission',
            element: guarded(<div style={{ padding: 24 }}>权限管理（开发中）</div>, 'permission:read'),
            meta: { title: '权限管理', permission: 'permission:read' },
          },
          {
            path: 'department',
            element: guarded(<div style={{ padding: 24 }}>部门管理（开发中）</div>, 'department:read'),
            meta: { title: '部门管理', permission: 'department:read' },
          },
          {
            path: 'audit',
            element: lazyGuarded(AuditList, 'audit:read'),
            meta: { title: '审计日志', permission: 'audit:read' },
          },
          {
            path: 'config',
            element: guarded(<div style={{ padding: 24 }}>系统配置（开发中）</div>, 'system:config'),
            meta: { title: '系统配置', permission: 'system:config' },
          },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <div style={{ padding: 50, textAlign: 'center' }}>404 - 页面不存在</div>,
    meta: { hidden: true },
  },
];

export default routes as RouteObject[];
