import React from 'react';
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { PUBLIC_CMS_KEYS } from '@/api/feature';
import FeatureEnabled from '@/components/FeatureEnabled/FeatureEnabled';
import { FeatureProvider, useFeature } from '@/hooks/useFeature';
import { getToken } from '@/utils/storage';
import styles from './PublicLayout.module.css';

const PublicNav: React.FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const { enabled } = useFeature('cms.public');
  const authed = Boolean(getToken());
  const loginHref = authed ? '/dashboard' : `/?next=${encodeURIComponent(location.pathname)}`;

  return (
    <header className={styles.header}>
      <Link to="/" className={styles.brand}>StarByte<span>.</span></Link>
      <nav className={styles.nav}>
        {enabled && (
          <>
            <NavLink to="/about-us" className={({ isActive }) => (isActive ? styles.active : undefined)}>
              {t('public.about')}
            </NavLink>
            <NavLink to="/docs" className={({ isActive }) => (isActive ? styles.active : undefined)}>
              {t('public.docs')}
            </NavLink>
          </>
        )}
        <Link to={loginHref}>{authed ? t('public.workbench') : t('public.login')}</Link>
      </nav>
    </header>
  );
};

const PublicLayout: React.FC = () => {
  const { t } = useTranslation();

  return (
    <FeatureProvider keys={PUBLIC_CMS_KEYS}>
      <div className={styles.shell}>
        <PublicNav />
        <main className={styles.main}>
          <FeatureEnabled flag="cms.public">
            <Outlet />
          </FeatureEnabled>
        </main>
        <footer className={styles.footer}>
          <span>{t('public.footer')}</span>
          <span>starbyte.smbu.edu.cn</span>
        </footer>
      </div>
    </FeatureProvider>
  );
};

export default PublicLayout;
