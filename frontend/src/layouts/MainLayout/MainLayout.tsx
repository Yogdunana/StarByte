import React, { useEffect } from 'react';
import { Layout, Menu, theme } from 'antd';
import type { MenuProps } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import * as Icons from '@ant-design/icons';

import TopBar from './components/TopBar';
import routes, { AppRouteObject } from '@/router/routes';
import { selectCollapsed } from '@/store/slices/appSlice';
import { fetchCurrentUser, selectCurrentUser } from '@/store/slices/userSlice';
import { getToken } from '@/utils/storage';
import type { AppDispatch } from '@/store';

const { Sider, Content } = Layout;

export interface MainLayoutProps {
  children?: React.ReactNode;
}

type MenuItem = NonNullable<MenuProps['items']>[number];

/** 递归构建菜单项 */
function getMenuItems(routeList: AppRouteObject[], parentPath = ''): MenuItem[] {
  return routeList
    .filter((route) => !route.meta?.hidden && route.path && route.path !== '*')
    .map((route) => {
      const fullPath = route.path!.startsWith('/')
        ? route.path!
        : `${parentPath}/${route.path}`;

      const IconComponent = route.meta?.icon
        ? (Icons as unknown as Record<string, React.FC>)[route.meta.icon]
        : null;

      const item = {
        key: fullPath,
        icon: IconComponent ? <IconComponent /> : null,
        label: route.meta?.title || route.path,
      } as MenuItem;

      if (route.children && route.children.length > 0) {
        const childItems = getMenuItems(
          route.children.filter((c) => c.path !== 'index'),
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

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

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

  const layoutRoute = (routes as AppRouteObject[]).find((r) => r.path === '/');
  const menuItems = layoutRoute?.children ? getMenuItems(layoutRoute.children, '') : [];

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
