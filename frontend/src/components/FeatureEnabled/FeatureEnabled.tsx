import React from 'react';
import { Empty } from 'antd';
import { useTranslation } from 'react-i18next';
import { useFeature } from '@/hooks/useFeature';

export interface FeatureEnabledProps {
  flag: string;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/** FeatureEnabled — 未命中灰度时渲染空态，不打断登录/CAS。 */
const FeatureEnabled: React.FC<FeatureEnabledProps> = ({ flag, children, fallback }) => {
  const { t } = useTranslation();
  const { enabled, loading } = useFeature(flag);
  if (loading) return null;
  if (!enabled) {
    return <>{fallback ?? <Empty description={t('feature.gated')} />}</>;
  }
  return <>{children}</>;
};

export default FeatureEnabled;
