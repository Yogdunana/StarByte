import React from 'react';
import { Layout, Avatar, Dropdown, Breadcrumb, theme } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  ProfileOutlined,
  BgColorsOutlined,
  GlobalOutlined,
} from '@ant-design/icons';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { toggleCollapsed } from '@/store/slices/appSlice';
import { selectCurrentUser, clearUser } from '@/store/slices/userSlice';
import { logout as logoutAction } from '@/store/slices/authSlice';
import { clearNotifications } from '@/store/slices/notificationSlice';
import { logout as logoutApi } from '@/api/auth';
import { removeToken } from '@/utils/storage';
import { useNotificationWebSocket } from '@/hooks/useNotificationWebSocket';
import NotificationBell from '@/components/NotificationBell';
import { useThemeLang } from '@/theme/ThemeLangContext';
import type { AppDispatch } from '@/store';

const { Header: AntHeader } = Layout;

const TopBar: React.FC = () => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const location = useLocation();
  const collapsed = useSelector((state: { app: { collapsed: boolean } }) => state.app.collapsed);
  const currentUser = useSelector(selectCurrentUser);
  const { setLang, setPreference, preference, lang } = useThemeLang();

  useNotificationWebSocket();

  const handleLogout = async () => {
    try {
      await logoutApi();
    } catch {
      // ignore
    }
    dispatch(logoutAction());
    dispatch(clearUser());
    dispatch(clearNotifications());
    removeToken();
    navigate('/login', { replace: true });
  };

  const userMenuItems = [
    { key: 'profile', icon: <ProfileOutlined />, label: t('topbar.profile'), onClick: () => navigate('/user/profile') },
    { key: 'settings', icon: <SettingOutlined />, label: t('topbar.settings'), onClick: () => navigate('/user/settings') },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: t('topbar.logout'), onClick: handleLogout },
  ];

  const themeItems = [
    { key: 'light', label: t('topbar.themeLight'), onClick: () => setPreference('light') },
    { key: 'dark', label: t('topbar.themeDark'), onClick: () => setPreference('dark') },
    { key: 'system', label: t('topbar.themeSystem'), onClick: () => setPreference('system') },
  ];

  const langItems = [
    { key: 'zh-CN', label: '简体中文', onClick: () => setLang('zh-CN') },
    { key: 'en-US', label: 'English', onClick: () => setLang('en-US') },
  ];

  const paths = location.pathname.split('/').filter(Boolean);
  const crumbs = paths.map((_, idx) => {
    const full = `/${paths.slice(0, idx + 1).join('/')}`;
    return { title: t(`menu.${full}`, { defaultValue: paths[idx] }) };
  });

  return (
    <AntHeader
      style={{
        padding: '0 16px',
        background: token.colorBgContainer,
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
        <Breadcrumb items={crumbs} />
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <Dropdown menu={{ items: themeItems, selectedKeys: [preference] }} placement="bottomRight">
          <BgColorsOutlined style={{ fontSize: 16, cursor: 'pointer' }} title={t('topbar.theme')} />
        </Dropdown>
        <Dropdown menu={{ items: langItems, selectedKeys: [lang] }} placement="bottomRight">
          <GlobalOutlined style={{ fontSize: 16, cursor: 'pointer' }} title={t('topbar.language')} />
        </Dropdown>
        <NotificationBell />
        <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
          <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer', gap: 8 }}>
            <Avatar size="small" src={currentUser?.avatar_url} icon={!currentUser?.avatar_url && <UserOutlined />} />
            <span style={{ fontSize: 14, display: 'flex', flexDirection: 'column', lineHeight: 1.2 }}>
              <span>{currentUser?.real_name || currentUser?.username || t('common.user')}</span>
              {currentUser?.student_no ? (
                <span style={{ fontSize: 12, color: token.colorTextSecondary }}>{currentUser.student_no}</span>
              ) : null}
            </span>
          </div>
        </Dropdown>
      </div>
    </AntHeader>
  );
};

export default TopBar;
