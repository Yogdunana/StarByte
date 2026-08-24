import React, { useEffect } from 'react';
import { Layout, Menu, theme } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import * as Icons from '@ant-design/icons';
import type { MenuProps } from 'antd';

import Header from './Header';
import routes, { AppRouteObject } from '@/router/routes';
import { selectCollapsed } from '@/store/slices/appSlice';
import { fetchCurrentUser, selectCurrentUser } from '@/store/slices/userSlice';
import { getToken } from '@/utils/storage';

const { Sider, Content } = Layout;

// 递归获取菜单项
const getMenuItems = (routeList: AppRouteObject[], parentPath = ''): MenuProps['items'] => {
  return routeList
    .filter((route) => !route.meta?.hidden && route.path && route.path !== '*')
    .map((route) => {
      const fullPath = route.path.startsWith('/')
        ? route.path
        : `${parentPath}/${route.path}`;

      const IconComponent = route.meta?.icon
        ? (Icons as any)[route.meta.icon]
        : null;

      const item: any = {
        key: fullPath,
        icon: IconComponent ? <IconComponent /> : null,
        label: route.meta?.title || route.path,
      };

      if (route.children && route.children.length > 0) {
        const childItems = getMenuItems(
          route.children.filter((c) => c.path !== 'index'),
          fullPath
        );
        if (childItems.length > 0) {
          item.children = childItems;
        }
      }

      return item;
    });
};

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch<any>();
  const collapsed = useSelector(selectCollapsed);
  const currentUser = useSelector(selectCurrentUser);

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // 检查登录状态并获取用户信息
  useEffect(() => {
    const token = getToken();
    if (!token) {
      navigate('/login', { replace: true });
      return;
    }

    if (!currentUser) {
      dispatch(fetchCurrentUser());
    }
  }, [dispatch, navigate, currentUser]);

  // 构建菜单数据
  const layoutRoute = (routes as AppRouteObject[]).find((r) => r.path === '/');
  const menuItems = layoutRoute?.children ? getMenuItems(layoutRoute.children, '') : [];

  // 计算当前选中的菜单项
  const getSelectedKeys = (): string[] => {
    const pathname = location.pathname;
    return [pathname];
  };

  // 计算展开的菜单项
  const getOpenKeys = (): string[] => {
    const pathname = location.pathname;
    const parts = pathname.split('/').filter(Boolean);
    if (parts.length < 2) return [];
    return [`/${parts[0]}`];
  };

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key);
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} width={220}>
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
          }}
        >
          {collapsed ? 'SB' : 'StarByte'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Layout>
        <Header />
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
