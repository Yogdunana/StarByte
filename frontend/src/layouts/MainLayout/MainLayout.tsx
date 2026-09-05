import React, { useEffect } from 'react';
import { Layout, Menu, Input, theme } from 'antd';
import type { MenuProps } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { SearchOutlined } from '@ant-design/icons';

import TopBar from './components/TopBar';
import { useMenu } from '@/hooks/useMenu';
import { fetchCurrentUser, selectCurrentUser } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';

const { Sider, Content } = Layout;

export interface MainLayoutProps {
  children?: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();
  const currentUser = useSelector(selectCurrentUser);
  const {
    menuItems,
    selectedKeys,
    openKeys,
    setOpenKeys,
    collapsed,
    searchKeyword,
    setSearchKeyword,
  } = useMenu();

  const {
    token: { colorBgContainer },
  } = theme.useToken();

  useEffect(() => {
    if (!currentUser) {
      dispatch(fetchCurrentUser());
    }
  }, [dispatch, currentUser]);

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
        {!collapsed && (
          <Input
            allowClear
            placeholder="搜索菜单"
            prefix={<SearchOutlined />}
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            style={{ margin: '8px 12px', width: 'calc(100% - 24px)' }}
          />
        )}
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          openKeys={collapsed ? [] : openKeys}
          onOpenChange={setOpenKeys}
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
            borderRadius: 16,
            boxShadow: '0 10px 30px rgba(15, 23, 42, 0.04)',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
