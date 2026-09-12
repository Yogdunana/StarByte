import React, { useState } from 'react';
import { Drawer, Grid, Input, Layout, Menu } from 'antd';
import { AppstoreOutlined, BellOutlined, CheckCircleOutlined, SearchOutlined, UserOutlined } from '@ant-design/icons';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import TopBar from './components/TopBar';
import { useMenu } from '@/hooks/useMenu';
import styles from './MainLayout.module.css';

export interface MainLayoutProps { children?: React.ReactNode }
const MainLayout: React.FC<MainLayoutProps> = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const screens = Grid.useBreakpoint();
  const mobile = !screens.lg;
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { menuItems, selectedKeys, openKeys, setOpenKeys, collapsed, searchKeyword, setSearchKeyword } = useMenu();
  const compact = !mobile && collapsed;
  const navigation = (
    <>
      <NavLink to="/dashboard" className={styles.brand} onClick={() => setDrawerOpen(false)} aria-label={t('shell.brandAria')}>
        <span className={styles.brandMark} aria-hidden="true"><span /><span /><span /><span /></span>
        {!compact && (
          <span>
            <strong>StarByte<span className={styles.brandDot}>.</span></strong>
            <small>{t('shell.subtitle')}</small>
          </span>
        )}
      </NavLink>
      {!compact && (
        <div className={styles.menuSearch}>
          <Input
            aria-label={t('common.searchMenu')}
            allowClear
            prefix={<SearchOutlined />}
            placeholder={t('common.searchMenu')}
            value={searchKeyword}
            onChange={(event) => setSearchKeyword(event.target.value)}
          />
        </div>
      )}
      <nav aria-label={t('shell.nav')} className={styles.menuArea}>
        <Menu
          mode="inline"
          inlineCollapsed={compact}
          selectedKeys={selectedKeys}
          openKeys={compact ? [] : openKeys}
          onOpenChange={setOpenKeys}
          items={menuItems}
          onClick={({ key }) => { navigate(key); setDrawerOpen(false); }}
          style={{ border: 0 }}
        />
      </nav>
      {!compact && (
        <div className={styles.sidebarFooter}>
          <span>{t('shell.motto')}</span>
          <small>{t('shell.orgEn')}</small>
        </div>
      )}
    </>
  );
  return (
    <Layout className={styles.shell}>
      <a href="#main-content" className={styles.skipLink}>{t('shell.skip')}</a>
      {!mobile && (
        <Layout.Sider
          trigger={null}
          collapsible
          collapsed={collapsed}
          width={248}
          collapsedWidth={76}
          className={styles.sidebar}
        >
          {navigation}
        </Layout.Sider>
      )}
      <Drawer
        title={t('shell.nav')}
        placement="left"
        open={mobile && drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={292}
        styles={{ body: { padding: 0, display: 'flex', flexDirection: 'column', background: 'var(--sb-surface)' } }}
      >
        {navigation}
      </Drawer>
      <Layout className={styles.main}>
        <TopBar mobile={mobile} onOpenMenu={() => setDrawerOpen(true)} />
        <Layout.Content id="main-content" tabIndex={-1} className={styles.content}><Outlet /></Layout.Content>
        <footer className={styles.footer}>
          <span>{t('shell.footerTag')}</span>
          <span>{t('shell.footerName')}</span>
        </footer>
      </Layout>
      {mobile && (
        <nav className={styles.mobileNav} aria-label={t('shell.quickNav')}>
          {[
            { path: '/dashboard', label: t('shell.workspace'), icon: <AppstoreOutlined /> },
            { path: '/task/my', label: t('shell.tasks'), icon: <CheckCircleOutlined /> },
            { path: '/notification/list', label: t('shell.messages'), icon: <BellOutlined /> },
            { path: '/user/profile', label: t('shell.me'), icon: <UserOutlined /> },
          ].map((item) => (
            <NavLink key={item.path} to={item.path} className={({ isActive }) => (isActive ? styles.mobileActive : '')}>
              {item.icon}
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
      )}
    </Layout>
  );
};
export default MainLayout;
