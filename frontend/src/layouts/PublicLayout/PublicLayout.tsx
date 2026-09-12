import React from 'react';
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getToken } from '@/utils/storage';
import styles from './PublicLayout.module.css';

const PublicLayout: React.FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const authed = Boolean(getToken());
  const loginHref = authed ? '/dashboard' : `/?next=${encodeURIComponent(location.pathname)}`;

  return (
    <div className={styles.shell}>
      <header className={styles.header}>
        <Link to="/" className={styles.brand}>StarByte<span>.</span></Link>
        <nav className={styles.nav}>
          <NavLink to="/about-us" className={({ isActive }) => (isActive ? styles.active : undefined)}>
            {t('public.about')}
          </NavLink>
          <NavLink to="/docs" className={({ isActive }) => (isActive ? styles.active : undefined)}>
            {t('public.docs')}
          </NavLink>
          <Link to={loginHref}>{authed ? t('public.workbench') : t('public.login')}</Link>
        </nav>
      </header>
      <main className={styles.main}>
        <Outlet />
      </main>
      <footer className={styles.footer}>
        <span>{t('public.footer')}</span>
        <span>starbyte.smbu.edu.cn</span>
      </footer>
    </div>
  );
};

export default PublicLayout;
