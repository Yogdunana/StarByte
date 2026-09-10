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
      <NavLink to="/dashboard" className={styles.brand} onClick={() => setDrawerOpen(false)} aria-label="StarByte 工作台">
        <span className={styles.brandMark} aria-hidden="true"><span /><span /><span /><span /></span>
        {!compact && <span><strong>StarByte<span className={styles.brandDot}>.</span></strong><small>计算机协会 · 协作空间</small></span>}
      </NavLink>
      {!compact && <div className={styles.menuSearch}>
        <Input aria-label={t('common.searchMenu')} allowClear prefix={<SearchOutlined />} placeholder={t('common.searchMenu')}
          value={searchKeyword} onChange={(event) => setSearchKeyword(event.target.value)} />
      </div>}
      <nav aria-label="主导航" className={styles.menuArea}>
        <Menu mode="inline" inlineCollapsed={compact} selectedKeys={selectedKeys}
          openKeys={compact ? [] : openKeys} onOpenChange={setOpenKeys} items={menuItems}
          onClick={({ key }) => { navigate(key); setDrawerOpen(false); }} style={{ border: 0 }} />
      </nav>
      {!compact && <div className={styles.sidebarFooter}><span>一起学习，一起创造。</span><small>COMPUTER ASSOCIATION</small></div>}
    </>
  );
  return (
    <Layout className={styles.shell}>
      <a href="#main-content" className={styles.skipLink}>跳转到主要内容</a>
      {!mobile && <Layout.Sider trigger={null} collapsible collapsed={collapsed} width={248} collapsedWidth={76} className={styles.sidebar}>{navigation}</Layout.Sider>}
      <Drawer title="导航" placement="left" open={mobile && drawerOpen} onClose={() => setDrawerOpen(false)} width={292}
        styles={{ body: { padding: 0, display: 'flex', flexDirection: 'column', background: 'var(--sb-bg-accent)' } }}>{navigation}</Drawer>
      <Layout className={styles.main}>
        <TopBar mobile={mobile} onOpenMenu={() => setDrawerOpen(true)} />
        <Layout.Content id="main-content" tabIndex={-1} className={styles.content}><Outlet /></Layout.Content>
        <footer className={styles.footer}><span>StarByte · 让协作井然有序</span><span>计算机协会管理平台</span></footer>
      </Layout>
      {mobile && <nav className={styles.mobileNav} aria-label="常用功能">
        {[
          { path: '/dashboard', label: '工作台', icon: <AppstoreOutlined /> },
          { path: '/task/my', label: '我的任务', icon: <CheckCircleOutlined /> },
          { path: '/notification/list', label: '消息', icon: <BellOutlined /> },
          { path: '/user/profile', label: '我的', icon: <UserOutlined /> },
        ].map(item => <NavLink key={item.path} to={item.path} className={({ isActive }) => isActive ? styles.mobileActive : ''}>{item.icon}<span>{item.label}</span></NavLink>)}
      </nav>}
    </Layout>
  );
};
export default MainLayout;
