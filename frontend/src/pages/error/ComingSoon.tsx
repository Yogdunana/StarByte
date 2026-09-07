import React from 'react';
import { useTranslation } from 'react-i18next';

const ComingSoon: React.FC<{ i18nKey: string }> = ({ i18nKey }) => {
  const { t } = useTranslation();
  return <div style={{ padding: 24 }}>{t(i18nKey)}</div>;
};

export default ComingSoon;
