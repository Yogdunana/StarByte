import React from 'react';
import { Layout, Avatar, Dropdown, Badge, Breadcrumb, Tooltip } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BellOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useLocation } from 'react-router-dom';

import { toggleCollapsed } from '@/store/slices/appSlice';
import { selectCurrentUser, clearUser } from '@/store/slices/userSlice';
import { logout as logoutAction } from '@/store/slices/authSlice';
import { logout as logoutApi } from '@/api/auth';
import { removeToken } from '@/utils/storage';
import type { AppDispatch } from '@/store';

const { Header: AntHeader } = Layout;

export interface TopBarProps {
  onToggleTheme?: () => void;
}

const TopBar: React.FC<TopBarProps> = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const location = useLocation();
  const collapsed = useSelector((state: { app: { collapsed: boolean } }) => state.app.collapsed);
  const currentUser = useSelector(selectCurrentUser);

  const handleLogout = async () => {
    try {
      await logoutApi();
    } catch {
      // 忽略登出 API 错误
    }
    dispatch(logoutAction());
    dispatch(clearUser());
    removeToken();
    navigate('/login', { replace: true });
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <ProfileOutlined />,
      label: '个人中心',
      onClick: () => navigate('/user/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '账号设置',
      onClick: () => navigate('/user/settings'),
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  const getBreadcrumbItems = () => {
    const paths = location.pathname.split('/').filter(Boolean);
    return paths.map((path) => ({
      title: path.charAt(0).toUpperCase() + path.slice(1),
    }));
  };

  return (
    <AntHeader
      style={{
        padding: '0 16px',
        background: '#fff',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        boxShadow: '0 1px 4px rgba(0,21,41,.08)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        {React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
          className: 'trigger',
          onClick: () => dispatch(toggleCollapsed()),
          style: { fontSize: 18, cursor: 'pointer' },
        })}
        <Breadcrumb items={getBreadcrumbItems()} />
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <Tooltip title="消息通知">
          <Badge count={0} size="small">
            <BellOutlined style={{ fontSize: 18, cursor: 'pointer' }} />
          </Badge>
        </Tooltip>

        <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
          <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer', gap: 8 }}>
            <Avatar
              size="small"
              src={currentUser?.avatar_url}
              icon={!currentUser?.avatar_url && <UserOutlined />}
            />
            <span style={{ fontSize: 14 }}>
              {currentUser?.real_name || currentUser?.username || '用户'}
            </span>
          </div>
        </Dropdown>
      </div>
    </AntHeader>
  );
};

export default TopBar;
