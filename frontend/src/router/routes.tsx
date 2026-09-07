import React, { lazy, Suspense } from 'react';
import { Navigate } from 'react-router-dom';
import type { RouteObject } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// 布局组件
import MainLayout from '@/layouts/MainLayout/MainLayout';

// 路由守卫
import AuthRoute from '@/router/guards/AuthRoute';
import PermissionRoute from '@/router/guards/PermissionRoute';
import ComingSoon from '@/pages/error/ComingSoon';

// 页面组件
const Login = lazy(() => import('@/pages/login/Login'));
const Dashboard = lazy(() => import('@/pages/dashboard/Dashboard'));
const BigScreenPage = lazy(() => import('@/pages/dashboard/bigscreen/BigScreenPage'));
const UserList = lazy(() => import('@/pages/user/UserList'));
const ProfileMePage = lazy(() => import('@/pages/user/ProfileMePage'));
const AccountSettingsPage = lazy(() => import('@/pages/user/AccountSettingsPage'));
const NotificationList = lazy(() => import('@/pages/notification/NotificationList'));
const TemplateList = lazy(() => import('@/pages/notification/TemplateList'));
const AuditList = lazy(() => import('@/pages/system/audit/AuditList'));
const DepartmentPage = lazy(() => import('@/pages/system/department/DepartmentPage'));
const ConfigPage = lazy(() => import('@/pages/system/config/ConfigPage'));
const DictPage = lazy(() => import('@/pages/system/dict/DictPage'));
const SessionPage = lazy(() => import('@/pages/system/session/SessionPage'));
const ExportPage = lazy(() => import('@/pages/system/export/ExportPage'));
const CachePage = lazy(() => import('@/pages/system/cache/CachePage'));
const SchedulerPage = lazy(() => import('@/pages/system/scheduler/SchedulerPage'));
const SearchPage = lazy(() => import('@/pages/system/search/SearchPage'));
const FileList = lazy(() => import('@/pages/file/FileList'));
const ApplicationPage = lazy(() => import('@/pages/member/application/ApplicationPage'));
const ProfilePage = lazy(() => import('@/pages/member/profile/ProfilePage'));
const InterviewSessionPage = lazy(() => import('@/pages/interview/SessionPage'));
const InterviewScorePage = lazy(() => import('@/pages/interview/ScorePage'));
const InterviewMyPage = lazy(() => import('@/pages/interview/MyInterviewPage'));
const InterviewStatsPage = lazy(() => import('@/pages/interview/StatsPage'));
const InterviewCheckinPage = lazy(() => import('@/pages/interview/CheckinPage'));
const MeetingListPage = lazy(() => import('@/pages/meeting/ListPage'));
const MeetingDetailPage = lazy(() => import('@/pages/meeting/DetailPage'));
const MeetingCheckinPage = lazy(() => import('@/pages/meeting/CheckinPage'));
const MeetingWeightPage = lazy(() => import('@/pages/meeting/WeightPage'));
const TaskListPage = lazy(() => import('@/pages/task/ListPage'));
const TaskBoardPage = lazy(() => import('@/pages/task/BoardPage'));
const TaskMyPage = lazy(() => import('@/pages/task/MyPage'));
const InternshipListPage = lazy(() => import('@/pages/internship/ListPage'));
const InternshipMyPage = lazy(() => import('@/pages/internship/MyPage'));
const InternshipStatsPage = lazy(() => import('@/pages/internship/StatsPage'));
const WorkflowDesigner = lazy(() => import('@/pages/workflow/designer/DesignerPage'));
const StatsOverviewPage = lazy(() => import('@/pages/stats/OverviewPage'));
const FormListPage = lazy(() => import('@/pages/form-designer/ListPage'));
const FormDesignerPage = lazy(() => import('@/pages/form-designer/DesignerPage'));
const FormFillPage = lazy(() => import('@/pages/form-designer/FillPage'));
const FormSubmissionsPage = lazy(() => import('@/pages/form-designer/SubmissionsPage'));
const FinancePage = lazy(() => import('@/pages/finance/FinancePage'));
const DisciplinePage = lazy(() => import('@/pages/discipline/DisciplinePage'));
const ContractPage = lazy(() => import('@/pages/contract/ContractPage'));
const Forbidden = lazy(() => import('@/pages/error/Forbidden'));
const NotFound = lazy(() => import('@/pages/error/NotFound'));

const LoadingFallback: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div style={{ padding: 24, textAlign: 'center' }}>{t('common.loading')}</div>
  );
};

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
    path: '/dashboard/bigscreen',
    element: (
      <AuthRoute>
        {lazyGuarded(BigScreenPage, 'stats:read')}
      </AuthRoute>
    ),
    meta: { title: '数据大屏', hidden: true, permission: 'stats:read' },
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
          {
            path: 'profile',
            element: lazyWrap(ProfileMePage),
            meta: { title: '个人中心', hidden: true },
          },
          {
            path: 'settings',
            element: lazyWrap(AccountSettingsPage),
            meta: { title: '账号设置', hidden: true },
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
            element: lazyWrap(ApplicationPage),
            meta: { title: '入会申请' },
          },
          {
            path: 'list',
            element: lazyGuarded(ProfilePage, 'member:read'),
            meta: { title: '会员档案', permission: 'member:read' },
          },
        ],
      },
      {
        path: 'interview',
        meta: { title: '面试管理', icon: 'ScheduleOutlined' },
        children: [
          {
            path: 'list',
            element: lazyGuarded(InterviewSessionPage, 'interview:read'),
            meta: { title: '面试安排', permission: 'interview:read' },
          },
          {
            path: 'score',
            element: lazyGuarded(InterviewScorePage, 'interview:evaluate'),
            meta: { title: '面试评分', permission: 'interview:evaluate' },
          },
          {
            path: 'my',
            element: lazyWrap(InterviewMyPage),
            meta: { title: '我的面试' },
          },
          {
            path: 'stats',
            element: lazyGuarded(InterviewStatsPage, 'interview:read'),
            meta: { title: '面试统计', permission: 'interview:read' },
          },
          {
            path: 'checkin',
            element: lazyWrap(InterviewCheckinPage),
            meta: { title: '面试签到' },
          },
        ],
      },
      {
        path: 'meeting',
        meta: { title: '会议管理', icon: 'CalendarOutlined' },
        children: [
          {
            path: 'list',
            element: lazyGuarded(MeetingListPage, 'meeting:read'),
            meta: { title: '会议列表', permission: 'meeting:read' },
          },
          {
            path: 'vote',
            element: lazyGuarded(MeetingWeightPage, 'meeting:read'),
            meta: { title: '投票权重', permission: 'meeting:read' },
          },
          {
            path: 'checkin',
            element: lazyWrap(MeetingCheckinPage),
            meta: { title: '会议签到', hidden: true },
          },
          {
            path: ':id',
            element: lazyGuarded(MeetingDetailPage, 'meeting:read'),
            meta: { title: '会议详情', permission: 'meeting:read', hidden: true },
          },
        ],
      },
      {
        path: 'task',
        meta: { title: '任务流转', icon: 'CheckCircleOutlined' },
        children: [
          {
            path: 'list',
            element: lazyGuarded(TaskListPage, 'task:read'),
            meta: { title: '任务列表', permission: 'task:read' },
          },
          {
            path: 'board',
            element: lazyGuarded(TaskBoardPage, 'task:read'),
            meta: { title: '任务看板', permission: 'task:read' },
          },
          {
            path: 'my',
            element: lazyWrap(TaskMyPage),
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
            element: <ComingSoon i18nKey="placeholder.workflowInstances" />,
            meta: { title: '流程实例' },
          },
          {
            path: 'todo',
            element: <ComingSoon i18nKey="placeholder.todo" />,
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
            element: lazyGuarded(InternshipListPage, 'internship:read'),
            meta: { title: '实习记录', permission: 'internship:read' },
          },
          {
            path: 'my',
            element: lazyWrap(InternshipMyPage),
            meta: { title: '我的实习' },
          },
          {
            path: 'stats',
            element: lazyGuarded(InternshipStatsPage, 'internship:read'),
            meta: { title: '实习统计', permission: 'internship:read' },
          },
        ],
      },
      {
        path: 'finance',
        element: lazyGuarded(FinancePage, 'finance:read'),
        meta: { title: '财务管理', icon: 'DollarOutlined', permission: 'finance:read' },
      },
      {
        path: 'discipline',
        element: lazyGuarded(DisciplinePage, 'discipline:read'),
        meta: { title: '纪律处分', icon: 'AlertOutlined', permission: 'discipline:read' },
      },
      {
        path: 'contract',
        element: lazyGuarded(ContractPage, 'contract:read'),
        meta: { title: '合同管理', icon: 'FileProtectOutlined', permission: 'contract:read' },
      },
      {
        path: 'forms',
        meta: { title: '动态表单', icon: 'FormOutlined' },
        children: [
          {
            index: true,
            element: lazyWrap(FormListPage),
            meta: { title: '表单列表' },
          },
          {
            path: 'designer',
            element: lazyGuarded(FormDesignerPage, 'form:write'),
            meta: { title: '表单设计', permission: 'form:write', hidden: true },
          },
          {
            path: 'designer/:id',
            element: lazyGuarded(FormDesignerPage, 'form:write'),
            meta: { title: '表单设计', permission: 'form:write', hidden: true },
          },
          {
            path: ':id/fill',
            element: lazyGuarded(FormFillPage, 'form:submit'),
            meta: { title: '填写表单', permission: 'form:submit', hidden: true },
          },
          {
            path: ':id/submissions',
            element: lazyGuarded(FormSubmissionsPage, 'form:read'),
            meta: { title: '提交记录', permission: 'form:read', hidden: true },
          },
        ],
      },
      {
        path: 'stats',
        meta: { title: '数据统计', icon: 'BarChartOutlined' },
        children: [
          {
            path: 'overview',
            element: lazyGuarded(StatsOverviewPage, 'stats:read'),
            meta: { title: '统计概览', permission: 'stats:read' },
          },
        ],
      },
      {
        path: 'files',
        element: lazyGuarded(FileList, 'file:read'),
        meta: { title: '文件管理', icon: 'FolderOutlined', permission: 'file:read' },
      },
      {
        path: 'system',
        meta: { title: '系统管理', icon: 'SettingOutlined' },
        children: [
          {
            path: 'role',
            element: guarded(<ComingSoon i18nKey="placeholder.roles" />, 'role:read'),
            meta: { title: '角色管理', permission: 'role:read' },
          },
          {
            path: 'permission',
            element: guarded(<ComingSoon i18nKey="placeholder.permissions" />, 'permission:read'),
            meta: { title: '权限管理', permission: 'permission:read' },
          },
          {
            path: 'department',
            element: lazyGuarded(DepartmentPage, 'department:read'),
            meta: { title: '部门管理', permission: 'department:read' },
          },
          {
            path: 'audit',
            element: lazyGuarded(AuditList, 'audit:read'),
            meta: { title: '审计日志', permission: 'audit:read' },
          },
          {
            path: 'dict',
            element: lazyGuarded(DictPage, 'dict:read'),
            meta: { title: '数据字典', permission: 'dict:read' },
          },
          {
            path: 'config',
            element: lazyGuarded(ConfigPage, 'config:read'),
            meta: { title: '系统配置', permission: 'config:read' },
          },
          {
            path: 'sessions',
            element: lazyGuarded(SessionPage, 'session:read'),
            meta: { title: '在线会话', permission: 'session:read' },
          },
          {
            path: 'export',
            element: lazyGuarded(ExportPage, 'export:read'),
            meta: { title: '打印导出', permission: 'export:read' },
          },
          {
            path: 'cache',
            element: lazyGuarded(CachePage, 'cache:read'),
            meta: { title: '缓存管理', permission: 'cache:read' },
          },
          {
            path: 'scheduler',
            element: lazyGuarded(SchedulerPage, 'scheduler:read'),
            meta: { title: '定时任务', permission: 'scheduler:read' },
          },
          {
            path: 'search',
            element: lazyGuarded(SearchPage, 'search:read'),
            meta: { title: '统一搜索', permission: 'search:read' },
          },
        ],
      },
      {
        path: '*',
        element: lazyWrap(NotFound),
        meta: { title: '页面不存在', hidden: true },
      },
    ],
  },
  {
    path: '*',
    element: lazyWrap(NotFound),
    meta: { title: '页面不存在', hidden: true },
  },
];

export default routes as RouteObject[];
