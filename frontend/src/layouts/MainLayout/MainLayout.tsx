import React, { useEffect } from 'react';
import { Layout, Menu, theme } from 'antd';
import type { MenuProps } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  ScheduleOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ApartmentOutlined,
  ReadOutlined,
  BarChartOutlined,
  SettingOutlined,
  BellOutlined,
} from '@ant-design/icons';

import TopBar from './components/TopBar';
import routes, { AppRouteObject, RouteMeta } from '@/router/routes';
import { selectCollapsed } from '@/store/slices/appSlice';
import { fetchCurrentUser, selectCurrentUser, selectPermissions } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';

const { Sider, Content } = Layout;

/** 路由 meta.icon 字符串到图标组件的类型安全映射 */
const iconMap: Record<string, React.FC> = {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  ScheduleOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ApartmentOutlined,
  ReadOutlined,
  BarChartOutlined,
  SettingOutlined,
  BellOutlined,
};

export interface MainLayoutProps {
  children?: React.ReactNode;
}

type MenuItem = NonNullable<MenuProps['items']>[number];

/**
 * 检查用户是否拥有指定权限
 * - permissions 包含 '*' 时视为超级管理员，拥有所有权限
 * - 未声明 permission 时视为公开，所有人可见
 */
function hasMenuPermission(permissions: string[], meta?: RouteMeta): boolean {
  if (permissions.includes('*')) return true;
  if (!meta?.permission) return true;
  return permissions.includes(meta.permission);
}

/**
 * 递归检查父路由下是否存在至少一个对当前用户可见的子路由
 */
function hasVisibleChildren(route: AppRouteObject, permissions: string[]): boolean {
  if (!route.children || route.children.length === 0) return false;
  return route.children.some((child) => {
    if (child.meta?.hidden) return false;
    if (child.children && child.children.length > 0) {
      return hasVisibleChildren(child, permissions);
    }
    return hasMenuPermission(permissions, child.meta);
  });
}

/**
 * 递归构建菜单项，按权限过滤
 */
function getMenuItems(routeList: AppRouteObject[], permissions: string[], parentPath = ''): MenuItem[] {
  return routeList
    .filter((route) => !route.meta?.hidden && route.path && route.path !== '*')
    .filter((route) => {
      // 父路由：至少有一个可见子路由时才显示
      if (route.children && route.children.length > 0) {
        return hasVisibleChildren(route, permissions);
      }
      // 叶子路由：检查自身权限
      return hasMenuPermission(permissions, route.meta);
    })
    .map((route) => {
      const fullPath = route.path!.startsWith('/')
        ? route.path!
        : `${parentPath}/${route.path}`;

      const IconComponent = route.meta?.icon
        ? iconMap[route.meta.icon]
        : null;

      const item = {
        key: fullPath,
        icon: IconComponent ? <IconComponent /> : null,
        label: route.meta?.title || route.path,
      } as MenuItem;

      if (route.children && route.children.length > 0) {
        const childItems = getMenuItems(
          route.children.filter((c) => c.path !== 'index'),
          permissions,
          fullPath,
        );
        if (childItems.length > 0) {
          (item as { children?: MenuItem[] }).children = childItems;
        }
      }

      return item;
    });
}

const MainLayout: React.FC<MainLayoutProps> = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch<AppDispatch>();
  const collapsed = useSelector(selectCollapsed);
  const currentUser = useSelector(selectCurrentUser);
  const permissions = useSelector(selectPermissions);

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // AuthRoute 已处理认证逻辑，这里仅作为安全兜底
  useEffect(() => {
    if (!currentUser) {
      dispatch(fetchCurrentUser());
    }
  }, [dispatch, currentUser]);

  const layoutRoute = (routes as AppRouteObject[]).find((r) => r.path === '/');
  const menuItems = layoutRoute?.children
    ? getMenuItems(layoutRoute.children, permissions, '')
    : [];

  const selectedKeys = [location.pathname];
  const parts = location.pathname.split('/').filter(Boolean);
  const openKeys = parts.length < 2 ? [] : [`/${parts[0]}`];

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key);
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={220}
        collapsedWidth={64}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontSize: collapsed ? 16 : 20,
            fontWeight: 'bold',
            borderBottom: '1px solid #2a3347',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
          }}
        >
          {collapsed ? 'SB' : 'StarByte'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          defaultOpenKeys={openKeys}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Layout>
        <TopBar />
        <Content
          style={{
            margin: 16,
            padding: 24,
            minHeight: 280,
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
