import { tx, useLocale } from '@/i18n/text';
import React from 'react';
import { Layout, Avatar, Dropdown, Breadcrumb, Button, theme } from 'antd';
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
import AnnouncementBadge from '@/components/AnnouncementBadge';
import { useThemeLang } from '@/theme/ThemeLangContext';
import type { AppDispatch } from '@/store';
import styles from './TopBar.module.css';

const { Header: AntHeader } = Layout;

interface TopBarProps {
  mobile?: boolean;
  onOpenMenu?: () => void;
}

const TopBar: React.FC<TopBarProps> = ({ mobile, onOpenMenu }) => {
  useLocale();
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const location = useLocation();
  const collapsed = useSelector((state: { app: { collapsed: boolean } }) => state.app.collapsed);
  const currentUser = useSelector(selectCurrentUser);
  const { setLang, setPreference, preference, lang, reduceMotion, setReduceMotion } =
    useThemeLang();

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
    navigate('/', { replace: true });
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <ProfileOutlined />,
      label: t('topbar.profile'),
      onClick: () => navigate('/user/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: t('topbar.settings'),
      onClick: () => navigate('/user/settings'),
    },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: t('topbar.logout'), onClick: handleLogout },
  ];

  const themeItems = [
    { key: 'light', label: t('topbar.themeLight'), onClick: () => setPreference('light') },
    { key: 'dark', label: t('topbar.themeDark'), onClick: () => setPreference('dark') },
    { key: 'system', label: t('topbar.themeSystem'), onClick: () => setPreference('system') },
    { type: 'divider' as const },
    {
      key: 'motion',
      label: reduceMotion ? t('topbar.restoreMotion') : t('topbar.reduceMotion'),
      onClick: () => setReduceMotion(!reduceMotion),
    },
  ];

  const langItems = [
    { key: 'zh-CN', label: tx('简体中文'), onClick: () => setLang('zh-CN') },
    { key: 'en-US', label: 'English', onClick: () => setLang('en-US') },
    { key: 'ru-RU', label: 'Русский', onClick: () => setLang('ru-RU') },
  ];

  const paths = location.pathname.split('/').filter(Boolean);
  const crumbs = paths.map((_, idx) => {
    const full = `/${paths.slice(0, idx + 1).join('/')}`;
    return {
      title: t(`menu.${full}`, {
        defaultValue: /^[0-9a-f]{8}-[0-9a-f-]{27}$/i.test(paths[idx])
          ? t('shell.detail')
          : paths[idx],
      }),
    };
  });

  const menuLabel = mobile
    ? t('shell.openNav')
    : collapsed
      ? t('shell.expandNav')
      : t('shell.collapseNav');

  return (
    <AntHeader className={styles.header}>
      <div className={styles.leading}>
        <Button
          type="text"
          aria-label={menuLabel}
          icon={collapsed || mobile ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={() => (mobile ? onOpenMenu?.() : dispatch(toggleCollapsed()))}
        />
        <Breadcrumb className={styles.breadcrumb} items={crumbs} />
      </div>

      <div className={styles.actions}>
        <Dropdown
          trigger={['click']}
          menu={{ items: themeItems, selectedKeys: [preference] }}
          placement="bottomRight"
        >
          <Button type="text" aria-label={t('topbar.theme')} icon={<BgColorsOutlined />} />
        </Dropdown>
        <Dropdown
          trigger={['click']}
          menu={{ items: langItems, selectedKeys: [lang] }}
          placement="bottomRight"
        >
          <Button
            type="text"
            className={styles.language}
            aria-label={t('topbar.language')}
            icon={<GlobalOutlined />}
          />
        </Dropdown>
        <AnnouncementBadge />
        <NotificationBell />
        <Dropdown trigger={['click']} menu={{ items: userMenuItems }} placement="bottomRight">
          <button type="button" className={styles.account} aria-label={t('shell.accountMenu')}>
            <Avatar
              size="small"
              src={currentUser?.avatar_url}
              icon={!currentUser?.avatar_url && <UserOutlined />}
            />
            <span className={styles.accountText}>
              <span>{currentUser?.real_name || currentUser?.username || t('common.user')}</span>
              {currentUser?.student_no ? (
                <span className={styles.studentNo} style={{ color: token.colorTextSecondary }}>
                  {currentUser.student_no}
                </span>
              ) : null}
            </span>
          </button>
        </Dropdown>
      </div>
    </AntHeader>
  );
};

export default TopBar;
